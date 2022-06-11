package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-chess/dao/mysql"
	"go-chess/dao/redis"
	"go-chess/util"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type connection struct {
	ws            *websocket.Conn
	send          chan []byte
	limitNum      int
	forbiddenWord bool
	timeLog       int64
}

type message struct {
	data   []byte
	roomId string
	name   string
	conn   *connection
}

type hub struct {
	rooms       map[string]map[*connection]bool
	broadcast   chan message
	broadcastss chan message
	warnings    chan message
	register    chan message
	unregister  chan message
	kickoutroom chan message
	warnmsg     chan message
}

var h = hub{
	broadcast:   make(chan message),
	broadcastss: make(chan message),
	warnings:    make(chan message),
	warnmsg:     make(chan message),
	register:    make(chan message),
	unregister:  make(chan message),
	kickoutroom: make(chan message),
	rooms:       make(map[string]map[*connection]bool),
}

func serverWs(ctx *gin.Context) {
	err := ctx.Request.ParseForm()
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	roomId := ctx.Request.Form["room_id"][0]
	Iuuid, _ := ctx.Get("uuid")
	uuid := Iuuid.(string)
	num, err := redis.RoomNum(roomId)

	if num == 2 || num == -1 {
		log.Println("room is full or err")
		util.RespError(ctx, 400, "room is full")
		return
	}

	err = redis.AddRoom(roomId, uuid)
	name, err := mysql.SelectUserNameByUUId(uuid)

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	m := message{nil, roomId, name, c}

	h.register <- m

	go m.writePump()
	go m.readPump()
}

func ready(ctx *gin.Context) {
	Iuuid, _ := ctx.Get("uuid")
	uuid := Iuuid.(string)
	roomId := ctx.Param("room_id")
	flag, err := redis.IsInRoom(roomId, uuid)
	if err != nil {
		log.Println(err)
		util.RespError(ctx, 400, "judge in the room err")
		return
	}
	if !flag {
		util.RespErrorWithData(ctx, 400, "ready error", "you are not in the room")
		return
	}

	err, key := redis.ReadySet(roomId, uuid)
	if err != nil || key == 2 {
		log.Println("ready or cancel error:", err)
		util.RespError(ctx, 400, "ready or cancel ready error")
		return
	}

	if key == 1 {
		util.RespSuccessful(ctx, "ready successful")
		return
	} else if key == 0 {
		util.RespSuccessful(ctx, "cancel ready successful")
		return
	}
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		//跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (m message) readPump() {
	c := m.conn

	defer func() {
		h.unregister <- m
		_ = c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	err := c.ws.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	c.ws.SetPongHandler(func(string) error {
		err := c.ws.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		return nil
	})

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				fmt.Println("unexpected close error:", err)
			}
			fmt.Println("err:", err)
			break
		}
		go m.Limit(msg)
	}
}

// Limit 发言违规进行时间限制禁言,超过三次禁言
func (m message) Limit(msg []byte) {
	c := m.conn
	//查看是否禁言，并判断是否超过了禁言时间
	timeNow := time.Now().Unix()
	if timeNow-c.timeLog < 300 {
		h.warnmsg <- m
	}

	// 不合法信息3次，判断是否有不合法信息，没有进行信息发布
	if c.limitNum >= 3 {
		h.kickoutroom <- m
		log.Println("素质太低，给你踢出去")
		_ = c.ws.Close() //
	} else //没有超过三次，可以继续
	{
		baseStr := "死傻操" //违法字符
		testStr := string(msg[:])
		for _, word := range testStr {
			//遍历是否有违法字符
			res := strings.Contains(baseStr, string(word))
			if res == true {
				c.limitNum += 1
				c.forbiddenWord = true //禁言
				//记录禁言开始时间
				c.timeLog = time.Now().Unix()
				h.warnings <- m
				break
			}
		}
		// 不禁言，消息合法 可以发送
		if c.forbiddenWord != true {
			// 通过所有检查，进行广播

			m := message{msg, m.roomId, m.name, c}
			h.broadcast <- m
		}
	}
}

func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (m *message) writePump() {
	c := m.conn

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (h *hub) run() {
	for {
		select {
		case m := <-h.register:
			conns := h.rooms[m.roomId]
			if conns == nil {
				conns = make(map[*connection]bool)
				h.rooms[m.roomId] = conns
			}
			h.rooms[m.roomId][m.conn] = true

			for con := range conns {
				sysmsg := "系统消息：欢迎新伙伴" + m.name + "加入" + m.roomId + "聊天室！！！"
				data := []byte(sysmsg)
				select {
				case con.send <- data:
				}
			}

		case m := <-h.unregister: //断开链接
			conns := h.rooms[m.roomId]
			if conns != nil {
				if _, ok := conns[m.conn]; ok {
					delete(conns, m.conn) //删除链接
					close(m.conn.send)
					for con := range conns {
						delMsg := "系统消息：" + m.name + "离开了" + m.roomId + "聊天室"
						data := []byte(delMsg)
						select {
						case con.send <- data:
						}
						if len(conns) == 0 {
							delete(h.rooms, m.roomId)
						}
					}
				}
			}

		case m := <-h.kickoutroom: //3次不合法信息后，被踢出群聊
			conns := h.rooms[m.roomId]
			notice := "由于您多次发送不合法信息,已被踢出群聊！！！"
			select {
			case m.conn.send <- []byte(notice):
			}
			if conns != nil {
				if _, ok := conns[m.conn]; ok {
					delete(conns, m.conn)
					close(m.conn.send)
					if len(conns) == 0 {
						delete(h.rooms, m.roomId)
					}
				}
			}

		case m := <-h.warnings:
			conns := h.rooms[m.roomId]
			if conns != nil {
				if _, ok := conns[m.conn]; ok {
					notice := "警告:您发布不合法信息，将禁言5分钟，三次后将被踢出群聊！！！"
					select {
					case m.conn.send <- []byte(notice):
					}
				}
			}

		case m := <-h.warnmsg: //禁言中提示
			conns := h.rooms[m.roomId]
			if conns != nil {
				if _, ok := conns[m.conn]; ok {
					notice := "您还在禁言中,暂时不能发送信息！！！"
					select {
					case m.conn.send <- []byte(notice):
					}
				}
			}

		case m := <-h.broadcast: //传输群信息/房间信息
			conns := h.rooms[m.roomId]
			for con := range conns {
				if con == m.conn { //自己发送的信息，不用再发给自己
					continue
				}
				select {
				case con.send <- m.data:
				default:
					close(con.send)
					delete(conns, con)
					if len(conns) == 0 {
						delete(h.rooms, m.roomId)
					}
				}
			}

		case m := <-h.broadcastss: //传输全员广播信息
			for _, conns := range h.rooms {
				for con := range conns {
					if con == m.conn { //自己发送的信息，不用再发给自己
						continue
					}
					select {
					case con.send <- m.data:
					default:
						close(con.send)
						delete(conns, con)
						if len(conns) == 0 {
							delete(h.rooms, m.roomId)
						}
					}
				}
			}
		}
	}
}

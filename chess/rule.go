package chess

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"math/rand"
)

type RC4Struct struct {
	s    [256]int
	x, y int
}

func (r *RC4Struct) initZero() {
	j := 0
	for i := 0; i < 256; i++ {
		r.s[i] = i
	}
	for i := 0; i < 256; i++ {
		j = (j + r.s[i]) & 255
		r.s[i], r.s[j] = r.s[j], r.s[i]
	}
}

func (r *RC4Struct) nextByte() uint32 {
	r.x = (r.x + 1) & 255
	r.y = (r.y + r.s[r.x]) & 255
	r.s[r.x], r.s[r.y] = r.s[r.y], r.s[r.x]
	return uint32(r.s[(r.s[r.x]+r.s[r.y])&255])
}

func (r *RC4Struct) nextLong() uint32 {
	uc0 := r.nextByte()
	uc1 := r.nextByte()
	uc2 := r.nextByte()
	uc3 := r.nextByte()
	return uc0 + (uc1 << 8) + (uc2 << 16) + (uc3 << 24)
}

type ZobristStruct struct {
	dwKey   uint32
	dwLock0 uint32
	dwLock1 uint32
}

func (z *ZobristStruct) initZero() {
	z.dwKey, z.dwLock0, z.dwLock1 = 0, 0, 0
}

func (z *ZobristStruct) initRC4(rc4 *RC4Struct) {
	z.dwKey = rc4.nextLong()
	z.dwLock0 = rc4.nextLong()
	z.dwLock1 = rc4.nextLong()
}

func (z *ZobristStruct) xor1(zobr *ZobristStruct) {
	z.dwKey ^= zobr.dwKey
	z.dwLock0 ^= zobr.dwLock0
	z.dwLock1 ^= zobr.dwLock1
}

func (z *ZobristStruct) xor2(zobr1, zobr2 *ZobristStruct) {
	z.dwKey ^= zobr1.dwKey ^ zobr2.dwKey
	z.dwLock0 ^= zobr1.dwLock0 ^ zobr2.dwLock0
	z.dwLock1 ^= zobr1.dwLock1 ^ zobr2.dwLock1
}

type Zobrist struct {
	Player *ZobristStruct          //走子方
	Table  [14][256]*ZobristStruct //所有棋子
}

func (z *Zobrist) initZobrist() {
	rc4 := &RC4Struct{}
	rc4.initZero()
	z.Player.initRC4(rc4)
	for i := 0; i < 14; i++ {
		for j := 0; j < 256; j++ {
			z.Table[i][j] = &ZobristStruct{}
			z.Table[i][j].initRC4(rc4)
		}
	}
}

type MoveStruct struct {
	ucpcCaptured int  //是否吃子
	ucbCheck     bool //是否将军
	wmv          int  //走法
	dwKey        uint32
}

func (m *MoveStruct) set(mv, pcCaptured int, bCheck bool, dwKey uint32) {
	m.wmv = mv
	m.ucpcCaptured = pcCaptured
	m.ucbCheck = bCheck
	m.dwKey = dwKey
}

type PositionStruct struct {
	sdPlayer    int                   //轮到谁走，0=红方，1=黑方
	vlRed       int                   //红方的子力价值
	vlBlack     int                   //黑方的子力价值
	nDistance   int                   //距离根节点的步数
	nMoveNum    int                   //历史走法数
	ucpcSquares [256]int              //棋盘上的棋子
	mvsList     [MaxMoves]*MoveStruct //历史走法信息列表
	zobr        *ZobristStruct        //走子方zobrist校验码
	zobrist     *Zobrist              //所有棋子zobrist校验码
	search      *Search
}

func NewPositionStruct() *PositionStruct {
	p := &PositionStruct{
		zobr: &ZobristStruct{
			dwKey:   0,
			dwLock0: 0,
			dwLock1: 0,
		},
		zobrist: &Zobrist{
			Player: &ZobristStruct{
				dwKey:   0,
				dwLock0: 0,
				dwLock1: 0,
			},
		},
		search: &Search{},
	}
	if p == nil {
		return nil
	}

	for i := 0; i < MaxMoves; i++ {
		tmpMoveStruct := &MoveStruct{}
		p.mvsList[i] = tmpMoveStruct
	}

	for i := 0; i < HashSize; i++ {
		p.search.hashTable[i] = &HashItem{}
	}

	p.zobrist.initZobrist()
	return p
}

func (p *PositionStruct) loadBook() bool {
	file, err := os.Open("./res/book.dat")
	if err != nil {
		fmt.Print(err)
		return false
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	if reader == nil {
		return false
	}

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Print(err)
				return false
			}
		}
		tmpLine := string(line)
		tmpResult := strings.Split(tmpLine, ",")
		if len(tmpResult) == 3 {
			tmpItem := &BookItem{}
			tmpdwLock, err := strconv.ParseUint(tmpResult[0], 10, 32)
			if err != nil {
				fmt.Print(err)
				continue
			}
			tmpItem.dwLock = uint32(tmpdwLock)
			tmpwmv, err := strconv.ParseInt(tmpResult[1], 10, 32)
			if err != nil {
				fmt.Print(err)
				continue
			}
			tmpItem.wmv = int(tmpwmv)
			tmpwvl, err := strconv.ParseInt(tmpResult[2], 10, 32)
			if err != nil {
				fmt.Print(err)
				continue
			}
			tmpItem.wvl = int(tmpwvl)

			p.search.BookTable = append(p.search.BookTable, tmpItem)
		}
	}
	return true
}

func (p *PositionStruct) clearBoard() {
	p.sdPlayer, p.vlRed, p.vlBlack, p.nDistance = 0, 0, 0, 0
	for i := 0; i < 256; i++ {
		p.ucpcSquares[i] = 0
	}
	p.zobr.initZero()
}

func (p *PositionStruct) setIrrev() {
	p.mvsList[0].set(0, 0, p.checked(), p.zobr.dwKey)
	p.nMoveNum = 1
}

func (p *PositionStruct) startup() {
	p.clearBoard()
	pc := 0
	for sq := 0; sq < 256; sq++ {
		pc = cucpcStartup[sq]
		if pc != 0 {
			p.addPiece(sq, pc)
		}
	}
	p.setIrrev()
}

func (p *PositionStruct) changeSide() {
	p.sdPlayer = 1 - p.sdPlayer
	p.zobr.xor1(p.zobrist.Player)
}

func (p *PositionStruct) addPiece(sq, pc int) {
	p.ucpcSquares[sq] = pc
	if pc < 16 {
		p.vlRed += cucvlPiecePos[pc-8][sq]
		p.zobr.xor1(p.zobrist.Table[pc-8][sq])
	} else {
		p.vlBlack += cucvlPiecePos[pc-16][squareFlip(sq)]
		p.zobr.xor1(p.zobrist.Table[pc-9][sq])
	}
}

func (p *PositionStruct) delPiece(sq, pc int) {
	p.ucpcSquares[sq] = 0
	if pc < 16 {
		p.vlRed -= cucvlPiecePos[pc-8][sq]
		p.zobr.xor1(p.zobrist.Table[pc-8][sq])
	} else {
		p.vlBlack -= cucvlPiecePos[pc-16][squareFlip(sq)]
		p.zobr.xor1(p.zobrist.Table[pc-9][sq])
	}
}

func (p *PositionStruct) evaluate() int {
	if p.sdPlayer == 0 {
		return p.vlRed - p.vlBlack + AdvancedValue
	}

	return p.vlBlack - p.vlRed + AdvancedValue
}

func (p *PositionStruct) inCheck() bool {
	return p.mvsList[p.nMoveNum-1].ucbCheck
}

func (p *PositionStruct) captured() bool {
	return p.mvsList[p.nMoveNum-1].ucpcCaptured != 0
}

func (p *PositionStruct) movePiece(mv int) int {
	sqSrc := src(mv)
	sqDst := dst(mv)
	pcCaptured := p.ucpcSquares[sqDst]
	if pcCaptured != 0 {
		p.delPiece(sqDst, pcCaptured)
	}
	pc := p.ucpcSquares[sqSrc]
	p.delPiece(sqSrc, pc)
	p.addPiece(sqDst, pc)
	return pcCaptured
}

func (p *PositionStruct) undoMovePiece(mv, pcCaptured int) {
	sqSrc := src(mv)
	sqDst := dst(mv)
	pc := p.ucpcSquares[sqDst]
	p.delPiece(sqDst, pc)
	p.addPiece(sqSrc, pc)
	if pcCaptured != 0 {
		p.addPiece(sqDst, pcCaptured)
	}
}

func (p *PositionStruct) makeMove(mv int) bool {
	dwKey := p.zobr.dwKey
	pcCaptured := p.movePiece(mv)
	if p.checked() {
		p.undoMovePiece(mv, pcCaptured)
		return false
	}
	p.changeSide()
	p.mvsList[p.nMoveNum].set(mv, pcCaptured, p.checked(), dwKey)
	p.nMoveNum++
	p.nDistance++
	return true
}

func (p *PositionStruct) undoMakeMove() {
	p.nDistance--
	p.nMoveNum--
	p.changeSide()
	p.undoMovePiece(p.mvsList[p.nMoveNum].wmv, p.mvsList[p.nMoveNum].ucpcCaptured)
}

func (p *PositionStruct) nullMove() {
	dwKey := p.zobr.dwKey
	p.changeSide()
	p.mvsList[p.nMoveNum].set(0, 0, false, dwKey)
	p.nMoveNum++
	p.nDistance++
}

func (p *PositionStruct) undoNullMove() {
	p.nDistance--
	p.nMoveNum--
	p.changeSide()
}

func (p *PositionStruct) nullOkay() bool {
	if p.sdPlayer == 0 {
		return p.vlRed > NullMargin
	}
	return p.vlBlack > NullMargin
}

func (p *PositionStruct) generateMoves(mvs []int, bCapture bool) int {
	nGenMoves, pcSrc, sqDst, pcDst, nDelta := 0, 0, 0, 0, 0
	pcSelfSide := sideTag(p.sdPlayer)
	pcOppSide := oppSideTag(p.sdPlayer)

	for sqSrc := 0; sqSrc < 256; sqSrc++ {
		if !inBoard(sqSrc) {
			continue
		}

		pcSrc = p.ucpcSquares[sqSrc]
		if (pcSrc & pcSelfSide) == 0 {
			continue
		}

		switch pcSrc - pcSelfSide {
		case PieceJiang:
			for i := 0; i < 4; i++ {
				sqDst = sqSrc + ccJiangDelta[i]
				if !inFort(sqDst) {
					continue
				}
				pcDst = p.ucpcSquares[sqDst]
				if (bCapture && (pcDst&pcOppSide) != 0) || (!bCapture && (pcDst&pcSelfSide) == 0) {
					mvs[nGenMoves] = move(sqSrc, sqDst)
					nGenMoves++
				}
			}
			break
		case PieceShi:
			for i := 0; i < 4; i++ {
				sqDst = sqSrc + ccShiDelta[i]
				if !inFort(sqDst) {
					continue
				}
				pcDst = p.ucpcSquares[sqDst]
				if (bCapture && (pcDst&pcOppSide) != 0) || (!bCapture && (pcDst&pcSelfSide) == 0) {
					mvs[nGenMoves] = move(sqSrc, sqDst)
					nGenMoves++
				}
			}
			break
		case PieceXiang:
			for i := 0; i < 4; i++ {
				sqDst = sqSrc + ccShiDelta[i]
				if !(inBoard(sqDst) && noRiver(sqDst, p.sdPlayer) && p.ucpcSquares[sqDst] == 0) {
					continue
				}
				sqDst += ccShiDelta[i]
				pcDst = p.ucpcSquares[sqDst]
				if (bCapture && (pcDst&pcOppSide) != 0) || (!bCapture && (pcDst&pcSelfSide) == 0) {
					mvs[nGenMoves] = move(sqSrc, sqDst)
					nGenMoves++
				}
			}
			break
		case PieceMa:
			for i := 0; i < 4; i++ {
				sqDst = sqSrc + ccJiangDelta[i]
				if p.ucpcSquares[sqDst] != 0 {
					continue
				}
				for j := 0; j < 2; j++ {
					sqDst = sqSrc + ccMaDelta[i][j]
					if !inBoard(sqDst) {
						continue
					}
					pcDst = p.ucpcSquares[sqDst]
					if (bCapture && (pcDst&pcOppSide) != 0) || (!bCapture && (pcDst&pcSelfSide) == 0) {
						mvs[nGenMoves] = move(sqSrc, sqDst)
						nGenMoves++
					}
				}
			}
			break
		case PieceJu:
			for i := 0; i < 4; i++ {
				nDelta = ccJiangDelta[i]
				sqDst = sqSrc + nDelta
				for inBoard(sqDst) {
					pcDst = p.ucpcSquares[sqDst]
					if pcDst == 0 {
						if !bCapture {
							mvs[nGenMoves] = move(sqSrc, sqDst)
							nGenMoves++
						}
					} else {
						if (pcDst & pcOppSide) != 0 {
							mvs[nGenMoves] = move(sqSrc, sqDst)
							nGenMoves++
						}
						break
					}
					sqDst += nDelta
				}

			}
			break
		case PiecePao:
			for i := 0; i < 4; i++ {
				nDelta = ccJiangDelta[i]
				sqDst = sqSrc + nDelta
				for inBoard(sqDst) {
					pcDst = p.ucpcSquares[sqDst]
					if pcDst == 0 {
						if !bCapture {
							mvs[nGenMoves] = move(sqSrc, sqDst)
							nGenMoves++
						}
					} else {
						break
					}
					sqDst += nDelta
				}
				sqDst += nDelta
				for inBoard(sqDst) {
					pcDst = p.ucpcSquares[sqDst]
					if pcDst != 0 {
						if (pcDst & pcOppSide) != 0 {
							mvs[nGenMoves] = move(sqSrc, sqDst)
							nGenMoves++
						}
						break
					}
					sqDst += nDelta
				}
			}
			break
		case PieceBing:
			sqDst = squareForward(sqSrc, p.sdPlayer)
			if inBoard(sqDst) {
				pcDst = p.ucpcSquares[sqDst]
				if (bCapture && (pcDst&pcOppSide) != 0) || (!bCapture && (pcDst&pcSelfSide) == 0) {
					mvs[nGenMoves] = move(sqSrc, sqDst)
					nGenMoves++
				}
			}
			if hasRiver(sqSrc, p.sdPlayer) {
				for nDelta = -1; nDelta <= 1; nDelta += 2 {
					sqDst = sqSrc + nDelta
					if inBoard(sqDst) {
						pcDst = p.ucpcSquares[sqDst]
						if (bCapture && (pcDst&pcOppSide) != 0) || (!bCapture && (pcDst&pcSelfSide) == 0) {
							mvs[nGenMoves] = move(sqSrc, sqDst)
							nGenMoves++
						}
					}
				}
			}
			break
		}
	}
	return nGenMoves
}

func (p *PositionStruct) legalMove(mv int) bool {
	sqSrc := src(mv)
	pcSrc := p.ucpcSquares[sqSrc]
	pcSelfSide := sideTag(p.sdPlayer)
	if (pcSrc & pcSelfSide) == 0 {
		return false
	}

	sqDst := dst(mv)
	pcDst := p.ucpcSquares[sqDst]
	if (pcDst & pcSelfSide) != 0 {
		return false
	}

	tmpPiece := pcSrc - pcSelfSide
	switch tmpPiece {
	case PieceJiang:
		return inFort(sqDst) && jiangSpan(sqSrc, sqDst)
	case PieceShi:
		return inFort(sqDst) && shiSpan(sqSrc, sqDst)
	case PieceXiang:
		return sameRiver(sqSrc, sqDst) && xiangSpan(sqSrc, sqDst) &&
			p.ucpcSquares[xiangPin(sqSrc, sqDst)] == 0
	case PieceMa:
		sqPin := maPin(sqSrc, sqDst)
		return sqPin != sqSrc && p.ucpcSquares[sqPin] == 0
	case PieceJu, PiecePao:
		nDelta := 0
		if sameX(sqSrc, sqDst) {
			if sqDst < sqSrc {
				nDelta = -1
			} else {
				nDelta = 1
			}
		} else if sameY(sqSrc, sqDst) {
			if sqDst < sqSrc {
				nDelta = -16
			} else {
				nDelta = 16
			}
		} else {
			return false
		}
		sqPin := sqSrc + nDelta
		for sqPin != sqDst && p.ucpcSquares[sqPin] == 0 {
			sqPin += nDelta
		}
		if sqPin == sqDst {
			return pcDst == 0 || tmpPiece == PieceJu
		} else if pcDst != 0 && tmpPiece == PiecePao {
			sqPin += nDelta
			for sqPin != sqDst && p.ucpcSquares[sqPin] == 0 {
				sqPin += nDelta
			}
			return sqPin == sqDst
		} else {
			return false
		}
	case PieceBing:
		if hasRiver(sqDst, p.sdPlayer) && (sqDst == sqSrc-1 || sqDst == sqSrc+1) {
			return true
		}
		return sqDst == squareForward(sqSrc, p.sdPlayer)
	default:

	}

	return false
}

func (p *PositionStruct) checked() bool {
	nDelta, sqDst, pcDst := 0, 0, 0
	pcSelfSide := sideTag(p.sdPlayer)
	pcOppSide := oppSideTag(p.sdPlayer)

	for sqSrc := 0; sqSrc < 256; sqSrc++ {
		if !inBoard(sqSrc) || p.ucpcSquares[sqSrc] != pcSelfSide+PieceJiang {
			continue
		}

		if p.ucpcSquares[squareForward(sqSrc, p.sdPlayer)] == pcOppSide+PieceBing {
			return true
		}
		for nDelta = -1; nDelta <= 1; nDelta += 2 {
			if p.ucpcSquares[sqSrc+nDelta] == pcOppSide+PieceBing {
				return true
			}
		}

		for i := 0; i < 4; i++ {
			if p.ucpcSquares[sqSrc+ccShiDelta[i]] != 0 {
				continue
			}
			for j := 0; j < 2; j++ {
				pcDst = p.ucpcSquares[sqSrc+ccMaCheckDelta[i][j]]
				if pcDst == pcOppSide+PieceMa {
					return true
				}
			}
		}

		for i := 0; i < 4; i++ {
			nDelta = ccJiangDelta[i]
			sqDst = sqSrc + nDelta
			for inBoard(sqDst) {
				pcDst = p.ucpcSquares[sqDst]
				if pcDst != 0 {
					if pcDst == pcOppSide+PieceJu || pcDst == pcOppSide+PieceJiang {
						return true
					}
					break
				}
				sqDst += nDelta
			}
			sqDst += nDelta
			for inBoard(sqDst) {
				pcDst = p.ucpcSquares[sqDst]
				if pcDst != 0 {
					if pcDst == pcOppSide+PiecePao {
						return true
					}
					break
				}
				sqDst += nDelta
			}
		}
		return false
	}
	return false
}

func (p *PositionStruct) isMate() bool {
	pcCaptured := 0
	mvs := make([]int, MaxGenMoves)
	nGenMoveNum := p.generateMoves(mvs, false)
	for i := 0; i < nGenMoveNum; i++ {
		pcCaptured = p.movePiece(mvs[i])
		if !p.checked() {
			p.undoMovePiece(mvs[i], pcCaptured)
			return false
		}

		p.undoMovePiece(mvs[i], pcCaptured)
	}
	return true
}

func (p *PositionStruct) drawValue() int {
	if p.nDistance&1 == 0 {
		return -DrawValue
	}

	return DrawValue
}

func (p *PositionStruct) repStatus(nRecur int) int {
	bSelfSide, bPerpCheck, bOppPerpCheck := false, true, true
	lpmvs := [MaxMoves]*MoveStruct{}
	for i := 0; i < MaxMoves; i++ {
		lpmvs[i] = p.mvsList[i]
	}

	for i := p.nMoveNum - 1; i >= 0 && lpmvs[i].wmv != 0 && lpmvs[i].ucpcCaptured == 0; i-- {
		if bSelfSide {
			bPerpCheck = bPerpCheck && lpmvs[i].ucbCheck
			if lpmvs[i].dwKey == p.zobr.dwKey {
				nRecur--
				if nRecur == 0 {
					result := 1
					if bPerpCheck {
						result += 2
					}
					if bOppPerpCheck {
						result += 4
					}
					return result
				}
			}
		} else {
			bOppPerpCheck = bOppPerpCheck && lpmvs[i].ucbCheck
		}
		bSelfSide = !bSelfSide
	}
	return 0
}

func (p *PositionStruct) repValue(nRepStatus int) int {
	vlReturn := 0
	if nRepStatus&2 != 0 {
		vlReturn += p.nDistance - BanValue
	}
	if nRepStatus&4 != 0 {
		vlReturn += BanValue - p.nDistance
	}

	if vlReturn == 0 {
		return p.drawValue()
	}

	return vlReturn
}

func (p *PositionStruct) mirror(posMirror *PositionStruct) {
	pc := 0
	posMirror.clearBoard()
	for sq := 0; sq < 256; sq++ {
		pc = p.ucpcSquares[sq]
		if pc != 0 {
			posMirror.addPiece(mirrorSquare(sq), pc)
		}
	}
	if p.sdPlayer == 1 {
		posMirror.changeSide()
	}
	posMirror.setIrrev()
}

type HashItem struct {
	ucDepth   int
	ucFlag    int
	svl       int
	wmv       int
	dwLock0   uint32
	dwLock1   uint32
	wReserved int
}

type BookItem struct {
	dwLock uint32
	wmv    int
	wvl    int
}

type Search struct {
	mvResult      int
	nHistoryTable [65536]int
	mvKillers     [LimitDepth][2]int
	hashTable     [HashSize]*HashItem
	BookTable     []*BookItem
}

func (p *PositionStruct) searchBook() int {
	bkToSearch := &BookItem{}
	mvs := make([]int, MaxGenMoves)
	vls := make([]int, MaxGenMoves)

	bookSize := len(p.search.BookTable)
	if bookSize <= 0 {
		return 0
	}

	bMirror := false
	bkToSearch.dwLock = p.zobr.dwLock1
	lpbk := sort.Search(bookSize, func(i int) bool {
		return p.search.BookTable[i].dwLock >= bkToSearch.dwLock
	})

	if lpbk == bookSize || (lpbk < bookSize && p.search.BookTable[lpbk].dwLock != bkToSearch.dwLock) {
		bMirror = true
		posMirror := NewPositionStruct()
		p.mirror(posMirror)
		bkToSearch.dwLock = posMirror.zobr.dwLock1
		lpbk = sort.Search(bookSize, func(i int) bool {
			return p.search.BookTable[i].dwLock >= bkToSearch.dwLock
		})
	}
	if lpbk == bookSize || (lpbk < bookSize && p.search.BookTable[lpbk].dwLock != bkToSearch.dwLock) {
		return 0
	}

	for lpbk >= 0 && p.search.BookTable[lpbk].dwLock == bkToSearch.dwLock {
		lpbk--
	}
	lpbk++

	vl, nBookMoves, mv := 0, 0, 0
	for lpbk < bookSize && p.search.BookTable[lpbk].dwLock == bkToSearch.dwLock {
		if bMirror {
			mv = mirrorMove(p.search.BookTable[lpbk].wmv)
		} else {
			mv = p.search.BookTable[lpbk].wmv
		}
		if p.legalMove(mv) {
			mvs[nBookMoves] = mv
			vls[nBookMoves] = p.search.BookTable[lpbk].wvl
			vl += vls[nBookMoves]
			nBookMoves++
			if nBookMoves == MaxGenMoves {
				break
			}
		}
		lpbk++
	}
	if vl == 0 {
		//防止"book.dat"中含有异常数据
		return 0
	}
	vl = rand.Intn(vl)
	i := 0
	for i = 0; i < nBookMoves; i++ {
		vl -= vls[i]
		if vl < 0 {
			break
		}
	}
	return mvs[i]
}
func (p *PositionStruct) probeHash(vlAlpha, vlBeta, nDepth int) (int, int) {
	hsh := p.search.hashTable[p.zobr.dwKey&(HashSize-1)]
	if hsh.dwLock0 != p.zobr.dwLock0 || hsh.dwLock1 != p.zobr.dwLock1 {
		return -MateValue, 0
	}
	mv := hsh.wmv
	bMate := false
	if hsh.svl > WinValue {
		if hsh.svl < BanValue {
			//可能导致搜索的不稳定性，立刻退出，但最佳着法可能拿到
			return -MateValue, mv
		}
		hsh.svl -= p.nDistance
		bMate = true
	} else if hsh.svl < -WinValue {
		if hsh.svl > -BanValue {
			//同上
			return -MateValue, mv
		}
		hsh.svl += p.nDistance
		bMate = true
	}
	if hsh.ucDepth >= nDepth || bMate {
		if hsh.ucFlag == HashBeta {
			if hsh.svl >= vlBeta {
				return hsh.svl, mv
			}
			return -MateValue, mv
		} else if hsh.ucFlag == HashAlpha {
			if hsh.svl <= vlAlpha {
				return hsh.svl, mv
			}
			return -MateValue, mv
		}
		return hsh.svl, mv
	}
	return -MateValue, mv
}

func (p *PositionStruct) RecordHash(nFlag, vl, nDepth, mv int) {
	hsh := p.search.hashTable[p.zobr.dwKey&(HashSize-1)]
	if hsh.ucDepth > nDepth {
		return
	}
	hsh.ucFlag = nFlag
	hsh.ucDepth = nDepth
	if vl > WinValue {
		if mv == 0 && vl <= BanValue {
			return
		}
		hsh.svl = vl + p.nDistance
	} else if vl < -WinValue {
		if mv == 0 && vl >= -BanValue {
			return //同上
		}
		hsh.svl = vl - p.nDistance
	} else {
		hsh.svl = vl
	}
	hsh.wmv = mv
	hsh.dwLock0 = p.zobr.dwLock0
	hsh.dwLock1 = p.zobr.dwLock1
	p.search.hashTable[p.zobr.dwKey&(HashSize-1)] = hsh
}

func (p *PositionStruct) mvvLva(mv int) int {
	return (cucMvvLva[p.ucpcSquares[dst(mv)]] << 3) - cucMvvLva[p.ucpcSquares[src(mv)]]
}

type SortStruct struct {
	mvHash    int   //置换表走法
	mvKiller1 int   //杀手走法
	mvKiller2 int   //杀手走法
	nPhase    int   //当前阶段
	nIndex    int   //当前采用第几个走法
	nGenMoves int   //总共有几个走法
	mvs       []int //所有的走法
}

func (p *PositionStruct) initSort(mvHash int, s *SortStruct) {
	if s == nil {
		return
	}

	s.mvHash = mvHash
	s.mvKiller1 = p.search.mvKillers[p.nDistance][0]
	s.mvKiller2 = p.search.mvKillers[p.nDistance][1]
	s.nPhase = PhaseHash
}

func (p *PositionStruct) nextSort(s *SortStruct) int {
	if s == nil {
		return 0
	}

	switch s.nPhase {
	case PhaseHash:
		s.nPhase = PhaseKiller1
		if s.mvHash != 0 {
			return s.mvHash
		}
		fallthrough
	case PhaseKiller1:
		s.nPhase = PhaseKiller2
		if s.mvKiller1 != s.mvHash && s.mvKiller1 != 0 && p.legalMove(s.mvKiller1) {
			return s.mvKiller1
		}
		fallthrough
	case PhaseKiller2:
		s.nPhase = PhaseGenMoves
		if s.mvKiller2 != s.mvHash && s.mvKiller2 != 0 && p.legalMove(s.mvKiller2) {
			return s.mvKiller2
		}
		fallthrough
	case PhaseGenMoves:
		s.nPhase = PhaseRest
		s.nGenMoves = p.generateMoves(s.mvs, false)
		s.mvs = s.mvs[:s.nGenMoves]
		sort.Slice(s.mvs, func(a, b int) bool {
			return p.search.nHistoryTable[a] > p.search.nHistoryTable[b]
		})
		s.nIndex = 0
		fallthrough
	case PhaseRest:
		for s.nIndex < s.nGenMoves {
			mv := s.mvs[s.nIndex]
			s.nIndex++
			if mv != s.mvHash && mv != s.mvKiller1 && mv != s.mvKiller2 {
				return mv
			}
		}
	default:
	}

	return 0
}

func (p *PositionStruct) setBestMove(mv, nDepth int) {
	p.search.nHistoryTable[mv] += nDepth * nDepth
	if p.search.mvKillers[p.nDistance][0] != mv {
		p.search.mvKillers[p.nDistance][1] = p.search.mvKillers[p.nDistance][0]
		p.search.mvKillers[p.nDistance][0] = mv
	}
}

func (p *PositionStruct) searchQuiesc(vlAlpha, vlBeta int) int {
	nGenMoves := 0
	mvs := make([]int, MaxGenMoves)

	vl := p.repStatus(1)
	if vl != 0 {
		return p.repValue(vl)
	}

	if p.nDistance == LimitDepth {
		return p.evaluate()
	}

	vlBest := -MateValue
	if p.inCheck() {
		nGenMoves = p.generateMoves(mvs, false)
		mvs = mvs[:nGenMoves]
		sort.Slice(mvs, func(a, b int) bool {
			return p.search.nHistoryTable[a] > p.search.nHistoryTable[b]
		})
	} else {
		vl = p.evaluate()
		if vl > vlBest {
			vlBest = vl
			if vl >= vlBeta {
				return vl
			}
			if vl > vlAlpha {
				vlAlpha = vl
			}
		}

		nGenMoves = p.generateMoves(mvs, true)
		mvs = mvs[:nGenMoves]
		sort.Slice(mvs, func(a, b int) bool {
			return p.mvvLva(mvs[a]) > p.mvvLva(mvs[b])
		})
	}

	for i := 0; i < nGenMoves; i++ {
		if p.makeMove(mvs[i]) {
			vl = -p.searchQuiesc(-vlBeta, -vlAlpha)
			p.undoMakeMove()
			if vl > vlBest {

				vlBest = vl

				if vl >= vlBeta {
					//Beta截断
					return vl
				}

				if vl > vlAlpha {
					vlAlpha = vl
				}
			}
		}
	}

	if vlBest == -MateValue {
		return p.nDistance - MateValue
	}
	return vlBest
}

func (p *PositionStruct) searchFull(vlAlpha, vlBeta, nDepth int, bNoNull bool) int {
	vl, mvHash, nNewDepth := 0, 0, 0

	if nDepth <= 0 {
		return p.searchQuiesc(vlAlpha, vlBeta)
	}

	vl = p.repStatus(1)
	if vl != 0 {
		return p.repValue(vl)
	}

	if p.nDistance == LimitDepth {
		return p.evaluate()
	}

	vl, mvHash = p.probeHash(vlAlpha, vlBeta, nDepth)
	if vl > -MateValue {
		return vl
	}

	if !bNoNull && !p.inCheck() && p.nullOkay() {
		p.nullMove()
		vl = -p.searchFull(-vlBeta, 1-vlBeta, nDepth-NullDepth-1, true)
		p.undoNullMove()
		if vl >= vlBeta {
			return vl
		}
	}

	nHashFlag := HashAlpha
	vlBest := -MateValue
	mvBest := 0

	tmpSort := &SortStruct{
		mvs: make([]int, MaxGenMoves),
	}
	p.initSort(mvHash, tmpSort)

	for mv := p.nextSort(tmpSort); mv != 0; mv = p.nextSort(tmpSort) {
		if p.makeMove(mv) {
			if p.inCheck() {
				nNewDepth = nDepth
			} else {
				nNewDepth = nDepth - 1
			}
			if vlBest == -MateValue {
				vl = -p.searchFull(-vlBeta, -vlAlpha, nNewDepth, false)
			} else {
				vl = -p.searchFull(-vlAlpha-1, -vlAlpha, nNewDepth, false)
				if vl > vlAlpha && vl < vlBeta {
					vl = -p.searchFull(-vlBeta, -vlAlpha, nNewDepth, false)
				}
			}
			p.undoMakeMove()

			if vl > vlBest {
				vlBest = vl
				if vl >= vlBeta {
					nHashFlag = HashBeta
					mvBest = mv
					break
				}
				if vl > vlAlpha {
					nHashFlag = HashPV
					mvBest = mv
					vlAlpha = vl
				}
			}
		}
	}

	if vlBest == -MateValue {
		//如果是杀棋，就根据杀棋步数给出评价
		return p.nDistance - MateValue
	}
	p.RecordHash(nHashFlag, vlBest, nDepth, mvBest)
	if mvBest != 0 {
		p.setBestMove(mvBest, nDepth)
	}
	return vlBest
}

func (p *PositionStruct) searchRoot(nDepth int) int {
	vl, nNewDepth := 0, 0
	vlBest := -MateValue
	tmpSort := &SortStruct{
		mvs: make([]int, MaxGenMoves),
	}
	p.initSort(p.search.mvResult, tmpSort)
	for mv := p.nextSort(tmpSort); mv != 0; mv = p.nextSort(tmpSort) {
		if p.makeMove(mv) {
			if p.inCheck() {
				nNewDepth = nDepth
			} else {
				nNewDepth = nDepth - 1
			}
			if vlBest == -MateValue {
				vl = -p.searchFull(-MateValue, MateValue, nNewDepth, true)
			} else {
				vl = -p.searchFull(-vlBest-1, -vlBest, nNewDepth, false)
				if vl > vlBest {
					vl = -p.searchFull(-MateValue, -vlBest, nNewDepth, true)
				}
			}
			p.undoMakeMove()
			if vl > vlBest {
				vlBest = vl
				p.search.mvResult = mv
				if vlBest > -WinValue && vlBest < WinValue {
					vlBest += int(rand.Int31()&RandomMask) - int(rand.Int31()&RandomMask)
				}
			}
		}
	}
	p.RecordHash(HashPV, vlBest, nDepth, p.search.mvResult)
	p.setBestMove(p.search.mvResult, nDepth)
	return vlBest
}

func (p *PositionStruct) searchMain() {
	for i := 0; i < 65536; i++ {
		p.search.nHistoryTable[i] = 0
	}
	for i := 0; i < LimitDepth; i++ {
		for j := 0; j < 2; j++ {
			p.search.mvKillers[i][j] = 0
		}
	}
	for i := 0; i < HashSize; i++ {
		p.search.hashTable[i].ucDepth = 0
		p.search.hashTable[i].ucFlag = 0
		p.search.hashTable[i].svl = 0
		p.search.hashTable[i].wmv = 0
		p.search.hashTable[i].wReserved = 0
		p.search.hashTable[i].dwLock0 = 0
		p.search.hashTable[i].dwLock1 = 0
	}
	start := time.Now()
	p.nDistance = 0

	p.search.mvResult = p.searchBook()
	if p.search.mvResult != 0 {
		p.makeMove(p.search.mvResult)
		if p.repStatus(3) == 0 {
			p.undoMakeMove()
			return
		}
		p.undoMakeMove()
	}
	vl := 0
	mvs := make([]int, MaxGenMoves)
	nGenMoves := p.generateMoves(mvs, false)
	for i := 0; i < nGenMoves; i++ {
		if p.makeMove(mvs[i]) {
			p.undoMakeMove()
			p.search.mvResult = mvs[i]
			vl++
		}
	}
	if vl == 1 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= LimitDepth; i++ {
		vl = p.searchRoot(i)
		if vl > WinValue || vl < -WinValue {
			break
		}
		if time.Now().Sub(start).Milliseconds() > 1000 {
			break
		}
	}
}
func (p *PositionStruct) printBoard() {
	stdString := "\n"
	for i, v := range p.ucpcSquares {
		if (i+1)%16 == 0 {
			tmpString := fmt.Sprintf("%2d\n", v)
			stdString += tmpString
		} else {
			tmpString := fmt.Sprintf("%2d ", v)
			stdString += tmpString
		}
	}
	fmt.Print(stdString)
}

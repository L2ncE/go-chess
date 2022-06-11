package chess

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/wav"
)

type Game struct {
	sqSelected     int                   //选中的格子
	mvLast         int                   //上一步棋
	bFlipped       bool                  //是否翻转棋盘
	bGameOver      bool                  //是否游戏结束
	showValue      string                //显示内容
	images         map[int]*ebiten.Image //图片资源
	audios         map[int]*audio.Player //音效
	audioContext   *audio.Context        //音效器
	singlePosition *PositionStruct       //棋局单例
}

func NewGame() bool {
	game := &Game{
		images:         make(map[int]*ebiten.Image),
		audios:         make(map[int]*audio.Player),
		singlePosition: NewPositionStruct(),
	}
	if game == nil || game.singlePosition == nil {
		return false
	}

	var err error

	game.audioContext, err = audio.NewContext(48000)
	if err != nil {
		fmt.Print(err)
		return false
	}

	if ok := game.loadResource(); !ok {
		return false
	}

	game.singlePosition.loadBook()
	game.singlePosition.startup()

	ebiten.SetWindowSize(BoardWidth, BoardHeight)
	ebiten.SetWindowTitle("中国象棋")
	if err := ebiten.RunGame(game); err != nil {
		fmt.Print(err)
		return false
	}

	return true
}

func (g *Game) Update(screen *ebiten.Image) error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.bGameOver {
			g.bGameOver = false
			g.showValue = ""
			g.sqSelected = 0
			g.mvLast = 0
			g.singlePosition.startup()
		} else {
			x, y := ebiten.CursorPosition()
			x = Left + (x-BoardEdge)/SquareSize
			y = Top + (y-BoardEdge)/SquareSize
			g.clickSquare1(squareXY(x, y))
			g.clickSquare2(squareXY(x, y))
		}
	}

	g.drawBoard(screen)
	if g.bGameOver {
		g.messageBox(screen)
	}
	return nil
}

func (g *Game) Layout(int, int) (screenWidth int, screenHeight int) {
	return BoardWidth, BoardHeight
}

func (g *Game) loadResource() bool {
	for k, v := range resMap {
		if k >= MusicSelect {
			d, err := wav.Decode(g.audioContext, audio.BytesReadSeekCloser(v))
			if err != nil {
				fmt.Print(err)
				return false
			}
			player, err := audio.NewPlayer(g.audioContext, d)
			if err != nil {
				fmt.Print(err)
				return false
			}
			g.audios[k] = player
		} else {
			img, _, err := image.Decode(bytes.NewReader(v))
			if err != nil {
				fmt.Print(err)
				return false
			}
			ebitenImage, _ := ebiten.NewImageFromImage(img, ebiten.FilterDefault)
			g.images[k] = ebitenImage
		}
	}

	return true
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	//棋盘
	if v, ok := g.images[ImgChessBoard]; ok {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, 0)
		screen.DrawImage(v, op)
	}

	for x := Left; x <= Right; x++ {
		for y := Top; y <= Bottom; y++ {
			xPos, yPos := 0, 0
			if g.bFlipped {
				xPos = BoardEdge + (xFlip(x)-Left)*SquareSize
				yPos = BoardEdge + (yFlip(y)-Top)*SquareSize
			} else {
				xPos = BoardEdge + (x-Left)*SquareSize
				yPos = BoardEdge + (y-Top)*SquareSize
			}
			sq := squareXY(x, y)
			pc := g.singlePosition.ucpcSquares[sq]
			if pc != 0 {
				g.drawChess(xPos, yPos+5, screen, g.images[pc])
			}
			if sq == g.sqSelected || sq == src(g.mvLast) || sq == dst(g.mvLast) {
				g.drawChess(xPos, yPos, screen, g.images[ImgSelect])
			}
		}
	}
}

func (g *Game) drawChess(x, y int, screen, img *ebiten.Image) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

func (g *Game) clickSquare1(sq int) {
	pc := 0
	if g.bFlipped {
		pc = g.singlePosition.ucpcSquares[squareFlip(sq)]
	} else {
		pc = g.singlePosition.ucpcSquares[sq]
	}

	if (pc & sideTag(g.singlePosition.sdPlayer)) != 0 {
		//如果点击自己的棋子，那么直接选中
		g.sqSelected = sq
		g.playAudio()
	} else if g.sqSelected != 0 && !g.bGameOver {
		//如果点击的不是自己的棋子，但有棋子选中了(一定是自己的棋子)，那么走这个棋子
		mv := move(g.sqSelected, sq)
		if g.singlePosition.legalMove(mv) {
			if g.singlePosition.makeMove(mv) {
				g.mvLast = mv
				g.sqSelected = 0
				//检查重复局面
				vlRep := g.singlePosition.repStatus(3)
				if g.singlePosition.isMate() {
					//如果分出胜负，那么播放胜负的声音，并且弹出不带声音的提示框
					g.playAudio()
					g.showValue = "Your Win!"
					g.bGameOver = true
				} else if vlRep > 0 {
					vlRep = g.singlePosition.repValue(vlRep)
					if vlRep > WinValue {
						g.playAudio()
						g.showValue = "Your Lose!"
					} else {
						if vlRep < -WinValue {
							g.playAudio()
							g.showValue = "Your Win!"
						} else {
							g.playAudio()
							g.showValue = "Your Draw!"
						}
					}
					g.bGameOver = true
				} else if g.singlePosition.nMoveNum > 100 {
					g.playAudio()
					g.showValue = "Your Draw!"
					g.bGameOver = true
				} else {
					if g.singlePosition.checked() {
						g.playAudio()
					} else {
						if g.singlePosition.captured() {
							g.playAudio()
							g.singlePosition.setIrrev()
						} else {
							g.playAudio()
						}
					}
					g.clickSquare1(sq)
				}
			}
		}
	}
}

func (g *Game) playAudio() {
	return
}

func (g *Game) clickSquare2(sq int) {
	pc := 0
	if g.bFlipped {
		pc = g.singlePosition.ucpcSquares[squareFlip(sq)]
	} else {
		pc = g.singlePosition.ucpcSquares[sq]
	}

	if (pc & sideTag(g.singlePosition.sdPlayer)) != 0 {
		//如果点击自己的棋子，那么直接选中
		g.sqSelected = sq
		g.playAudio()
	} else if g.sqSelected != 0 && !g.bGameOver {
		//如果点击的不是自己的棋子，但有棋子选中了(一定是自己的棋子)，那么走这个棋子
		mv := move(g.sqSelected, sq)
		if g.singlePosition.legalMove(mv) {
			if g.singlePosition.makeMove(mv) {
				g.mvLast = mv
				g.sqSelected = 0
				//检查重复局面
				vlRep := g.singlePosition.repStatus(3)
				if g.singlePosition.isMate() {
					//如果分出胜负，那么播放胜负的声音，并且弹出不带声音的提示框
					g.playAudio()
					g.showValue = "Your Win!"
					g.bGameOver = true
				} else if vlRep > 0 {
					vlRep = g.singlePosition.repValue(vlRep)
					if vlRep > WinValue {
						g.playAudio()
						g.showValue = "Your Lose!"
					} else {
						if vlRep < -WinValue {
							g.playAudio()
							g.showValue = "Your Win!"
						} else {
							g.playAudio()
							g.showValue = "Your Draw!"
						}
					}
					g.bGameOver = true
				} else if g.singlePosition.nMoveNum > 100 {
					g.playAudio()
					g.showValue = "Your Draw!"
					g.bGameOver = true
				} else {
					if g.singlePosition.checked() {
						g.playAudio()
					} else {
						if g.singlePosition.captured() {
							g.playAudio()
							g.singlePosition.setIrrev()
						} else {
							g.playAudio()
						}
					}
				}
			} else {
				g.playAudio()
			}
		}
	}
}

func (g *Game) messageBox(screen *ebiten.Image) {
	fmt.Println(g.showValue)
	tt, err := truetype.Parse(fonts.ArcadeN_ttf)
	if err != nil {
		fmt.Print(err)
		return
	}
	arcadeFont := truetype.NewFace(tt, &truetype.Options{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	text.Draw(screen, g.showValue, arcadeFont, 180, 288, color.White)
	text.Draw(screen, "Click mouse to restart", arcadeFont, 100, 320, color.White)
}

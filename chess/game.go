package chess

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"log"
)

//Game 象棋窗口
type Game struct {
}

//NewGame 创建象棋程序
func NewGame() bool {
	game := &Game{}
	if game == nil {
		return false
	}

	ebiten.SetWindowSize(520, 576)
	ebiten.SetWindowTitle("中国象棋")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

func (g *Game) Layout(int, int) (screenWidth int, screenHeight int) {
	return 520, 576
}

func (g *Game) Update(screen *ebiten.Image) error {
	img, _, err := ebitenutil.NewImageFromFile("./static/ChessBoard.png", ebiten.FilterDefault)
	if err != nil {
		log.Print(err)
		return err
	}
	op := &ebiten.DrawImageOptions{}
	_ = screen.DrawImage(img, op)
	return nil
}

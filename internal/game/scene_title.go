package game

import (
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// titleScene 是标题画面：展示游戏名与最高分，等待玩家点击/按键开始。
type titleScene struct {
	ticks int
}

func newTitleScene() *titleScene {
	return &titleScene{}
}

// Update 只做一件事：等待任意一次「跳跃」输入，然后进入游戏。
func (s *titleScene) Update(g *Game) error {
	s.ticks++
	if g.input.JumpJustPressed() {
		g.switchTo(newPlayScene(g))
	}
	return nil
}

func (s *titleScene) Draw(g *Game, screen *ebiten.Image) {
	drawBackground(screen, g.assets.Sky)

	// 让地鼠轻轻地上下浮动，静止画面也显得有生气。
	bob := math.Sin(float64(s.ticks)/15) * 8
	size := g.assets.Gopher.Bounds()
	drawSprite(screen, g.assets.Gopher, (ScreenWidth-float64(size.Dx()))/2, 130+bob)

	drawTextCentered(screen, "FLAPPY GOPHER", g.fontLarge, 40, color.White)

	if g.best > 0 {
		drawTextCentered(screen, "BEST "+strconv.Itoa(g.best), g.fontSmall, 240, color.White)
	}

	// 提示文字闪烁：每 30 个 tick（约半秒）切换一次显示状态。
	if (s.ticks/30)%2 == 0 {
		drawTextCentered(screen, "CLICK OR PRESS SPACE", g.fontSmall, 290, color.White)
	}
}

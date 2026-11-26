package game

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// restartDelay 是失败后多久才开始接受「重新开始」的输入（单位：tick）。
// 玩家失误时往往还按着键，加一点延迟可以避免「还没看清分数就又开局了」。
const restartDelay = 20

// gameOverScene 是游戏结束画面。
// 它保留了对失败那一刻 playScene 的引用，用来把游戏世界「定格」在背景上，
// 让玩家看清楚自己是撞上了哪根管子。
type gameOverScene struct {
	prev  *playScene
	ticks int
}

func newGameOverScene(prev *playScene) *gameOverScene {
	return &gameOverScene{prev: prev}
}

func (s *gameOverScene) Update(g *Game) error {
	s.ticks++
	if s.ticks < restartDelay {
		return nil
	}
	if g.input.JumpJustPressed() {
		g.switchTo(newPlayScene(g))
	}
	return nil
}

func (s *gameOverScene) Draw(g *Game, screen *ebiten.Image) {
	// 先画定格的世界，再压暗，最后叠上文字。
	s.prev.drawWorld(g, screen)
	fillOverlay(screen, 150)

	score := s.prev.pipes.Score()

	drawTextCentered(screen, "GAME OVER", g.fontLarge, 80, color.White)
	drawTextCentered(screen, "SCORE "+strconv.Itoa(score), g.fontSmall, 140, color.White)
	drawTextCentered(screen, "BEST "+strconv.Itoa(g.best), g.fontSmall, 170, color.White)

	// 刷新纪录时给一点正反馈。
	if score > 0 && score == g.best {
		drawTextCentered(screen, "NEW BEST!", g.fontSmall, 200, color.RGBA{R: 0xff, G: 0xd7, A: 0xff})
	}

	if s.ticks >= restartDelay && (s.ticks/30)%2 == 0 {
		drawTextCentered(screen, "CLICK OR PRESS SPACE TO RETRY", g.fontSmall, 260, color.White)
	}
}

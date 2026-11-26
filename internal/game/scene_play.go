package game

import (
	"image/color"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// playScene 是真正的游戏画面：地鼠、管子、分数都在这里。
type playScene struct {
	player *Player
	pipes  *PipeField
}

// newPlayScene 开始新的一局。
func newPlayScene(g *Game) *playScene {
	size := g.assets.Gopher.Bounds()
	return &playScene{
		player: NewPlayer(float64(size.Dx()), float64(size.Dy())),
		// 用当前时间当随机种子，每一局管子的位置都不一样。
		pipes: NewPipeField(uint64(time.Now().UnixNano())),
	}
}

// Update 推进一个 tick 的游戏逻辑。
func (s *playScene) Update(g *Game) error {
	// 1. 处理输入：按一次跳一次。
	if g.input.JumpJustPressed() {
		s.player.Jump()
		g.sfx.Flap()
	}

	// 2. 推进物理模拟与场景卷动。顺序很重要：
	//    先让角色运动，再让管子运动并判定碰撞，保证两者用的是同一时刻的状态。
	s.player.Update()
	scoreBefore := s.pipes.Score()
	s.pipes.Update(s.player.X)
	if s.pipes.Score() > scoreBefore {
		g.sfx.Score()
	}

	// 3. 失败判定：撞到管子、撞到天花板、掉出屏幕底部。
	body := s.player.Rect()
	if s.pipes.Collides(body) || body.Y < 0 || body.Bottom() > ScreenHeight {
		// 先把角色贴回屏幕内，这样结束画面上「定格」的角色不会被裁掉一半。
		if body.Y < 0 {
			s.player.Y = 0
		}
		if body.Bottom() > ScreenHeight {
			s.player.Y = ScreenHeight - s.player.H
		}

		g.recordBest(s.pipes.Score())
		g.sfx.Hit()
		g.switchTo(newGameOverScene(s))
	}

	return nil
}

func (s *playScene) Draw(g *Game, screen *ebiten.Image) {
	s.drawWorld(g, screen)

	// 分数显示在顶部中央。
	drawTextCentered(screen, strconv.Itoa(s.pipes.Score()), g.fontLarge, 12, color.White)

	if g.debug {
		s.drawDebug(screen)
	}
}

// drawWorld 只绘制「游戏世界」本身（背景、管子、角色），不画 HUD。
// 游戏结束画面会复用它，把失败那一刻的画面定格在背景上。
func (s *playScene) drawWorld(g *Game, screen *ebiten.Image) {
	drawBackground(screen, g.assets.Sky)
	for _, p := range s.pipes.Pipes() {
		p.Draw(screen, g.assets.Wall)
	}
	s.player.Draw(screen, g.assets.Gopher)
}

// drawDebug 在调试模式下显示所有碰撞盒。
func (s *playScene) drawDebug(screen *ebiten.Image) {
	pipeColor := color.RGBA{R: 0xff, A: 0xff}
	playerColor := color.RGBA{G: 0xff, A: 0xff}

	for _, p := range s.pipes.Pipes() {
		strokeRect(screen, p.TopRect(), pipeColor)
		strokeRect(screen, p.BottomRect(), pipeColor)
	}
	strokeRect(screen, s.player.Rect(), playerColor)
}

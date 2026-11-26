package game

import (
	"image/color"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// playScene 是真正的游戏画面：地鼠、管子、道具、分数、血条都在这里。
type playScene struct {
	player *Player
	pipes  *PipeField
	health *Health
	rng    *rand.Rand // 道具生成用的随机数发生器
	ticks  int        // 场景已经运行了多少个 tick，用于制作闪烁动画
}

// newPlayScene 开始新的一局。
func newPlayScene(g *Game) *playScene {
	size := g.assets.Gopher.Bounds()

	// 用当前时间当随机种子，每一局管子的位置、掉落的道具都不一样。
	seed := uint64(time.Now().UnixNano())

	s := &playScene{
		player: NewPlayer(float64(size.Dx()), float64(size.Dy())),
		pipes:  NewPipeField(seed),
		health: NewHealth(MaxHearts),
		rng:    rand.New(rand.NewPCG(seed+1, seed>>1)),
	}

	// 道具规则要参考血条、角色大小等状态，所以由场景告诉管子区域「这根管子要不要带道具」。
	s.pipes.spawnItem = s.pickItem
	return s
}

// Update 推进一个 tick 的游戏逻辑。
func (s *playScene) Update(g *Game) error {
	s.ticks++
	s.health.Update()

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

	// 3. 拾取道具：道具浮在空隙中心，角色经过时自动拾取。
	//    先处理拾取再判定碰撞，这样刚捡到的头盔能立刻挡住同一 tick 的碰撞，
	//    缩小道具也会让本 tick 的碰撞盒立刻变小。
	for _, kind := range s.pipes.CollectItems(s.player.Rect()) {
		if sp := itemSpecOf(kind); sp != nil {
			sp.Apply(s)
		}
	}

	// 4. 受伤判定：撞到管子、顶到天花板、掉到地板都会扣一次血
	//    （有头盔时先消耗头盔的免疫次数，不扣心血）。
	body := s.player.Rect()
	hurt := s.pipes.Collides(body)
	if body.Y < 0 {
		// 把角色贴回屏幕内，否则角色会有一半留在画面外，
		// 而且结束画面上「定格」的角色也会被裁掉一半。
		s.player.Y = 0
		s.player.VelY = 0
		hurt = true
	}
	if body.Bottom() > ScreenHeight {
		s.player.Y = ScreenHeight - body.H
		s.player.VelY = 0
		hurt = true
	}

	// Hurt 在无敌时间内会返回 false：一次碰撞只会造成一次伤害，不会连续掉血。
	if hurt && s.health.Hurt() {
		if s.health.Dead() {
			// 最后一击不再播受伤音效，换成死亡音效，避免两个声音叠在一起。
			g.sfx.Die()
			g.recordBest(s.pipes.Score())
			g.switchTo(newGameOverScene(s))
		} else {
			g.sfx.Hurt()
		}
	}

	return nil
}

func (s *playScene) Draw(g *Game, screen *ebiten.Image) {
	s.drawWorld(g, screen)

	// 分数显示在顶部中央，血条（以及头盔）在左上角。
	drawTextCentered(screen, strconv.Itoa(s.pipes.Score()), g.fontLarge, 12, color.White)
	s.drawHearts(g, screen)
	s.drawShield(g, screen)

	if g.debug {
		s.drawDebug(screen)
	}
}

// drawWorld 只绘制「游戏世界」本身（背景、管子、道具、角色），不画 HUD。
// 游戏结束画面会复用它，把失败那一刻的画面定格在背景上。
func (s *playScene) drawWorld(g *Game, screen *ebiten.Image) {
	drawBackground(screen, g.assets.Sky)
	for _, p := range s.pipes.Pipes() {
		p.Draw(screen, g.assets.Wall)
		if it := p.Item(); it != nil {
			it.Draw(screen, itemImage(g.assets, it.Kind))
		}
	}
	s.player.Draw(screen, g.assets.Gopher, s.playerAlpha())
}

// playerAlpha 让角色在受伤后的无敌时间内闪烁，
// 提示玩家「这段时间里再撞到东西也不会掉血」。
func (s *playScene) playerAlpha() float64 {
	// 每 4 个 tick 切换一次透明度，比整段时间都半透明更容易被注意到。
	if !s.health.Dead() && s.health.Invulnerable() && (s.ticks/4)%2 == 0 {
		return 0.35
	}
	return 1
}

// drawHearts 在左上角画出血条：剩余的心用不透明图标，已经失去的用半透明图标。
func (s *playScene) drawHearts(g *Game, screen *ebiten.Image) {
	img := g.assets.Heart
	// 贴图是 1024x1024，按 HUD 需要的尺寸等比缩放。
	scale := HeartSize / float64(img.Bounds().Dy())

	for i := 0; i < s.health.Max(); i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(HeartMarginX+float64(i)*(HeartSize+HeartGap), HeartMarginY)
		if i >= s.health.Hearts() {
			// 空槽位画得淡一些，玩家一眼就能数出还剩几颗心。
			op.ColorScale.ScaleAlpha(0.25)
		}
		screen.DrawImage(img, op)
	}
}

// drawShield 把头盔画在血条后面：免疫次数越少图标越淡，用完就消失。
func (s *playScene) drawShield(g *Game, screen *ebiten.Image) {
	charges := s.health.ShieldCharges()
	if charges <= 0 {
		return
	}

	// 紧跟在最后一颗心后面。
	x := HeartMarginX + float64(s.health.Max())*(HeartSize+HeartGap)
	alpha := 0.25 + 0.75*float64(charges)/float64(MaxShield)
	drawSpriteScaled(screen, g.assets.ItemHelmet, x, HeartMarginY, HeartSize, alpha)
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

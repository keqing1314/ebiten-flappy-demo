package game

import (
	"math"
	"testing"

	"github.com/keqing1314/ebiten-flappy-demo/internal/sound"
)

func TestRectIntersects(t *testing.T) {
	base := Rect{X: 0, Y: 0, W: 10, H: 10}

	tests := []struct {
		name  string
		other Rect
		want  bool
	}{
		{"完全重合", Rect{X: 0, Y: 0, W: 10, H: 10}, true},
		{"部分重叠", Rect{X: 5, Y: 5, W: 10, H: 10}, true},
		{"被完全包含", Rect{X: 2, Y: 2, W: 3, H: 3}, true},
		{"右侧贴合不算碰撞", Rect{X: 10, Y: 0, W: 10, H: 10}, false},
		{"下方贴合不算碰撞", Rect{X: 0, Y: 10, W: 10, H: 10}, false},
		{"完全分离", Rect{X: 100, Y: 100, W: 10, H: 10}, false},
	}

	for _, tt := range tests {
		if got := base.Intersects(tt.other); got != tt.want {
			t.Errorf("%s: Intersects() = %v, 期望 %v", tt.name, got, tt.want)
		}
	}
}

func TestPlayerJumpThenFall(t *testing.T) {
	p := NewPlayer(60, 75)
	if want := float64(ScreenHeight-75) / 2; p.Y != want {
		t.Fatalf("初始 y = %v, 期望 %v", p.Y, want)
	}

	startY := p.Y
	p.Jump()
	if p.VelY != JumpVelocity {
		t.Fatalf("跳跃后 VelY = %v, 期望 %v", p.VelY, JumpVelocity)
	}

	p.Update()
	if p.Y >= startY {
		t.Fatalf("跳跃后应当向上移动, y = %v, 起点 = %v", p.Y, startY)
	}

	// 一直不操作，下落速度会加速到上限并保持不变。
	for i := 0; i < 300; i++ {
		p.Update()
	}
	if p.VelY != MaxFallSpeed {
		t.Fatalf("下落速度应被限制为 %v, 实际 %v", MaxFallSpeed, p.VelY)
	}
	if p.Y <= startY {
		t.Fatalf("长时间不操作后应当落回下方, y = %v", p.Y)
	}
}

func TestPipeFieldSpawn(t *testing.T) {
	f := NewPipeField(42)
	if len(f.Pipes()) != 1 {
		t.Fatalf("开局应当生成 1 根管子, 实际 %d", len(f.Pipes()))
	}

	p := f.Pipes()[0]
	if p.X != ScreenWidth {
		t.Errorf("新管子应当从屏幕右边缘进入, X = %v", p.X)
	}
	if p.GapY < PipeGapMargin {
		t.Errorf("空隙上边缘 %v 越过了上边距 %v", p.GapY, PipeGapMargin)
	}
	if bottom := p.GapY + PipeGapHeight; bottom > ScreenHeight-PipeGapMargin {
		t.Errorf("空隙下边缘 %v 越过了下边距", bottom)
	}
}

func TestPipeFieldSpawnInterval(t *testing.T) {
	f := NewPipeField(1)

	// 关掉计分干扰：这里只关心生成节奏。
	for i := 0; i < PipeInterval-1; i++ {
		f.Update(0)
	}
	if n := len(f.Pipes()); n != 1 {
		t.Fatalf("%d 个 tick 后应有 1 根管子, 实际 %d", PipeInterval-1, n)
	}

	f.Update(0)
	if n := len(f.Pipes()); n != 2 {
		t.Fatalf("%d 个 tick 后应有 2 根管子, 实际 %d", PipeInterval, n)
	}
}

func TestPipeFieldScoresOncePerPipe(t *testing.T) {
	f := NewPipeField(7)
	f.timer = math.MaxInt // 关掉自动生成, 只留下开局那根管子

	for i := 0; i < 1000; i++ {
		f.Update(PlayerX)
	}

	if got := f.Score(); got != 1 {
		t.Fatalf("同一根管子只应记 1 分, 实际 %d", got)
	}
	if n := len(f.Pipes()); n != 0 {
		t.Fatalf("移出屏幕的管子应当被回收, 还剩 %d 根", n)
	}
}

func TestPipeFieldGapChangeIsLimited(t *testing.T) {
	f := NewPipeField(9)
	center := func(p *Pipe) float64 { return p.GapY + PipeGapHeight/2 }
	prev := center(f.Pipes()[0])

	for i := 0; i < 30; i++ {
		f.pipes = nil // 只关心新生成的这根管子
		f.spawn()
		cur := center(f.pipes[0])
		if math.Abs(cur-prev) > MaxGapDelta {
			t.Fatalf("相邻管子空隙中心变化 %v，超过上限 %v", math.Abs(cur-prev), MaxGapDelta)
		}
		prev = cur
	}
}

func TestHealthHurtAndInvulnerability(t *testing.T) {
	h := NewHealth(MaxHearts)
	if h.Hearts() != MaxHearts || h.Max() != MaxHearts || h.Dead() {
		t.Fatalf("满血状态不正确: hearts=%d max=%d", h.Hearts(), h.Max())
	}

	if !h.Hurt() {
		t.Fatal("满血时第一次碰撞应当扣血")
	}
	if got := h.Hearts(); got != MaxHearts-1 {
		t.Fatalf("受伤后剩余 %d 颗心, 期望 %d", got, MaxHearts-1)
	}
	if !h.Invulnerable() {
		t.Fatal("受伤后应当进入无敌时间")
	}

	// 无敌时间内无论被判定多少次碰撞，都只算一次伤害。
	for i := 0; i < InvulnTicks-1; i++ {
		h.Update()
		if h.Hurt() {
			t.Fatalf("无敌时间内不应再次扣血（第 %d tick）", i+1)
		}
	}
	if !h.Invulnerable() {
		t.Fatal("倒计时未走完时仍应处于无敌状态")
	}

	h.Update()
	if h.Invulnerable() {
		t.Fatal("经过 InvulnTicks 个 tick 后无敌时间应当结束")
	}
	if !h.Hurt() {
		t.Fatal("无敌时间结束后应当可以再次扣血")
	}
	if got := h.Hearts(); got != MaxHearts-2 {
		t.Fatalf("第二次受伤后剩余 %d 颗心, 期望 %d", got, MaxHearts-2)
	}
}

func TestHealthDiesWhenHeartsRunOut(t *testing.T) {
	h := NewHealth(2)

	for i := 0; i < 2; i++ {
		if !h.Hurt() {
			t.Fatalf("第 %d 次受伤应当生效", i+1)
		}
		for j := 0; j < InvulnTicks; j++ {
			h.Update()
		}
	}

	if !h.Dead() {
		t.Fatal("两颗心都扣完后应当判定为死亡")
	}
	if got := h.Hearts(); got != 0 {
		t.Fatalf("死亡时剩余 %d 颗心, 期望 0", got)
	}
	if h.Hurt() {
		t.Fatal("死亡后不应再扣血")
	}
}

// newTestPlayScene 创建一个没有管子干扰的游戏画面，只用来测试天花板/地板受伤逻辑。
func newTestPlayScene(hearts int) *playScene {
	s := &playScene{
		player: NewPlayer(60, 75),
		pipes:  NewPipeField(1),
		health: NewHealth(hearts),
	}
	s.pipes.pipes = nil
	s.pipes.timer = math.MaxInt
	return s
}

// newTestGame 创建一个只带音效的游戏实例（音效播放器全是 nil，播放是空操作）。
func newTestGame() *Game {
	return &Game{sfx: &sound.Sound{}}
}

func TestPlaySceneCeilingHurtsAndClamps(t *testing.T) {
	s := newTestPlayScene(3)
	s.player.Y = 1
	s.player.VelY = -5

	if err := s.Update(newTestGame()); err != nil {
		t.Fatal(err)
	}

	if got := s.health.Hearts(); got != 2 {
		t.Fatalf("撞到天花板后剩余 %d 颗心, 期望 2", got)
	}
	if s.player.Y != 0 {
		t.Fatalf("角色应被贴回屏幕顶边, y = %v", s.player.Y)
	}
	if s.player.VelY != 0 {
		t.Fatalf("撞到天花板后竖直速度应清零, VelY = %v", s.player.VelY)
	}
}

func TestPlaySceneFloorDrainsHeartsThenGameOver(t *testing.T) {
	s := newTestPlayScene(3)
	g := newTestGame()

	for i := 0; i < 600 && s.health.Hearts() == 3; i++ {
		if err := s.Update(g); err != nil {
			t.Fatal(err)
		}
		if s.player.Rect().Bottom() > ScreenHeight {
			t.Fatalf("角色不应跑出屏幕底部, y = %v", s.player.Y)
		}
	}
	if got := s.health.Hearts(); got != 2 {
		t.Fatalf("掉到地板后剩余 %d 颗心, 期望 2", got)
	}

	// 一直不操作，角色会一直贴着地板，但无敌时间保证每 60 tick 才扣一次血。
	for i := 0; i < 600; i++ {
		if _, over := g.scene.(*gameOverScene); over {
			break
		}
		if err := s.Update(g); err != nil {
			t.Fatal(err)
		}
	}

	if !s.health.Dead() {
		t.Fatal("血量应当已经耗尽")
	}
	if _, ok := g.scene.(*gameOverScene); !ok {
		t.Fatalf("血量耗尽后应当切换到结束画面, 实际场景类型 %T", g.scene)
	}
}

func TestPipeFieldCollides(t *testing.T) {
	f := NewPipeField(3)
	f.timer = math.MaxInt
	p := f.Pipes()[0]

	tests := []struct {
		name string
		body Rect
		want bool
	}{
		{"撞到上半根管子", Rect{X: p.X + 1, Y: p.GapY - 10, W: 10, H: 10}, true},
		{"撞到下半根管子", Rect{X: p.X + 1, Y: p.GapY + PipeGapHeight + 5, W: 10, H: 10}, true},
		{"从空隙中穿过", Rect{X: p.X + 1, Y: p.GapY + 5, W: 10, H: 10}, false},
		{"还在管子左侧", Rect{X: p.X - 20, Y: p.GapY - 10, W: 10, H: 10}, false},
		{"在管子右侧", Rect{X: p.X + PipeWidth + 1, Y: p.GapY - 10, W: 10, H: 10}, false},
	}

	for _, tt := range tests {
		if got := f.Collides(tt.body); got != tt.want {
			t.Errorf("%s: Collides() = %v, 期望 %v", tt.name, got, tt.want)
		}
	}
}

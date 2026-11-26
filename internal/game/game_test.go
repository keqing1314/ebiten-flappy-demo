package game

import (
	"math"
	"testing"
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

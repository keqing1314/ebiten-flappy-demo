package game

import (
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// Pipe 是一根管子。虽然看起来是两段，但逻辑上它是一根「中间挖了个洞」的直管：
//
//	上半根占据 y ∈ [0, GapY]
//	下半根占据 y ∈ [GapY+PipeGapHeight, ScreenHeight]
//
// 碰撞盒直接用矩形，而不是贴图的逐像素轮廓——像素级碰撞又慢又容易被美术细节坑到，
// 用略小于贴图的矩形手感反而更好。
type Pipe struct {
	X    float64 // 管子左边缘的 x 坐标
	GapY float64 // 空隙上边缘的 y 坐标

	scored bool // 是否已经为这根管子加过分，保证一根管子只加一次
}

// TopRect 返回上半根管子的碰撞盒。
func (p *Pipe) TopRect() Rect {
	return Rect{X: p.X, Y: 0, W: PipeWidth, H: p.GapY}
}

// BottomRect 返回下半根管子的碰撞盒。
func (p *Pipe) BottomRect() Rect {
	y := p.GapY + PipeGapHeight
	return Rect{X: p.X, Y: y, W: PipeWidth, H: ScreenHeight - y}
}

// Draw 绘制上下两段管子。
//
// 贴图 wall.png 高 361px，比屏幕还高一点，所以贴在屏幕外侧多出来的部分会被自动裁掉，
// 效果就是一根从屏幕边缘伸出来的通顶管子。
func (p *Pipe) Draw(dst, wall *ebiten.Image) {
	h := float64(wall.Bounds().Dy())

	// 上半根：贴图的【下边缘】要对齐空隙上边缘，所以从 GapY 往上偏移一个贴图高度。
	drawSprite(dst, wall, p.X, p.GapY-h)

	// 下半根：贴图的【上边缘】对齐空隙下边缘。
	drawSprite(dst, wall, p.X, p.GapY+PipeGapHeight)
}

// ScrollSpeedAt 返回存活 ticks 个 tick 后的卷动速度（像素/tick）。
//
// 速度随存活时间按对数增长：前期提升最明显，之后逐渐放缓，最终停在 ScrollSpeedMax：
//
//	speed = min(ScrollSpeedMax, ScrollSpeed + ScrollSpeedGain*ln(1+ticks/ScrollSpeedRampTicks))
//
// 之所以用对数而不是线性，是因为线性增长迟早会让管子快到无法反应，
// 而对数曲线天然有上限，既能让玩家感到难度在涨，又不会失控。
func ScrollSpeedAt(ticks int) float64 {
	if ticks <= 0 {
		return ScrollSpeed
	}
	speed := ScrollSpeed + ScrollSpeedGain*math.Log1p(float64(ticks)/ScrollSpeedRampTicks)
	if speed > ScrollSpeedMax {
		return ScrollSpeedMax
	}
	return speed
}

// PipeField 负责管理屏幕上所有管子：生成、移动、计分、回收。
type PipeField struct {
	pipes      []*Pipe
	timer      int // 距离下一次生成还有多少 tick
	ticks      int // 已经存活了多少个 tick，用来计算随存活时间提升的卷动速度
	rng        *rand.Rand
	score      int
	lastCenter float64 // 上一根管子空隙中心的高度，用来限制变化幅度
}

// NewPipeField 创建一片管子区域，并立刻生成第一根。
//
// seed 用于初始化随机数发生器：固定 seed 就能复现同一串管子，
// 这让单元测试和「回放 bug」都变得可行。实际游戏里传入当前时间即可。
func NewPipeField(seed uint64) *PipeField {
	f := &PipeField{
		timer:      PipeInterval,
		rng:        rand.New(rand.NewPCG(seed, seed>>1)),
		lastCenter: (minGapCenter + maxGapCenter) / 2,
	}
	f.spawn()
	return f
}

// Score 返回当前得分（成功穿过的管子数量）。
func (f *PipeField) Score() int {
	return f.score
}

// Pipes 返回当前所有管子，供绘制与调试使用（调用方不应该修改切片内容）。
func (f *PipeField) Pipes() []*Pipe {
	return f.pipes
}

// Update 推进一个 tick：移动（速度随存活时间提升）→ 计分 → 回收 → 生成。
// playerX 是角色左边缘的 x 坐标，用来判断管子是否已经被甩到身后。
func (f *PipeField) Update(playerX float64) {
	f.ticks++
	speed := ScrollSpeedAt(f.ticks)

	for _, p := range f.pipes {
		p.X -= speed

		// 管子整体越过角色左边缘，说明玩家成功穿过，记 1 分（每根只记一次）。
		if !p.scored && p.X+PipeWidth < playerX {
			p.scored = true
			f.score++
		}
	}

	// 回收移出屏幕左侧的管子，否则切片会随游戏时长无限增长。
	// 这里复用底层数组（pipes[:0]），避免每次 Update 都分配新内存。
	alive := f.pipes[:0]
	for _, p := range f.pipes {
		if p.X+PipeWidth > 0 {
			alive = append(alive, p)
		}
	}
	f.pipes = alive

	// 到时间就生成下一根。
	f.timer--
	if f.timer <= 0 {
		f.spawn()
		f.timer = PipeInterval
	}
}

// Collides 判断给定矩形是否撞到了任意一根管子。
func (f *PipeField) Collides(r Rect) bool {
	for _, p := range f.pipes {
		if r.Intersects(p.TopRect()) || r.Intersects(p.BottomRect()) {
			return true
		}
	}
	return false
}

// spawn 在屏幕右边缘外侧生成一根新管子，空隙中心高度随机，
// 但会受上一根管子的高度约束，保证难度是「跳一跳够得着」的。
func (f *PipeField) spawn() {
	center := minGapCenter + f.rng.Float64()*(maxGapCenter-minGapCenter)

	// 先把高度差夹在 [lastCenter-MaxGapDelta, lastCenter+MaxGapDelta]，
	// 再夹回合法范围（上一步可能把它推出边界）。
	center = math.Min(math.Max(center, f.lastCenter-MaxGapDelta), f.lastCenter+MaxGapDelta)
	center = math.Min(math.Max(center, minGapCenter), maxGapCenter)
	f.lastCenter = center

	f.pipes = append(f.pipes, &Pipe{
		X:    ScreenWidth,
		GapY: center - PipeGapHeight/2,
	})
}

package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Player 是玩家控制的地鼠。
//
// 它用的是最经典的「速度 + 加速度」积分模型：
//   - 拍翅膀：直接给 VelY 赋一个向上的初速度（相当于施加一次冲量）；
//   - 每个 tick：VelY 先加上重力，再让 Y 加上 VelY。
//
// 这种写法比「直接修改位置」自然得多，是所有 2D 平台跳跃/飞行游戏的基础。
//
// 坐标 (X, Y) 是角色碰撞盒的左上角。
type Player struct {
	X, Y  float64
	W, H  float64 // 贴图原始尺寸；碰撞盒与绘制都会再乘上 Scale
	Scale float64 // 缩放比例，1 是原始大小（缩小道具会把它调小）
	VelY  float64 // 竖直速度，单位：像素/tick，向下为正

	shrinkTicks int // 缩小效果还剩多少 tick，0 表示没有缩小
}

// NewPlayer 在屏幕左侧、垂直居中的位置创建一个静止的角色。
// 尺寸取自贴图本身（gopher.png 是 60x75），这样换素材时不需要改代码。
func NewPlayer(w, h float64) *Player {
	return &Player{
		X:     PlayerX,
		Y:     (ScreenHeight - h) / 2,
		W:     w,
		H:     h,
		Scale: 1,
	}
}

// Shrink 让角色缩小，并在 ticks 个 tick 后自动恢复原尺寸（缩小道具的效果）。
// 缩小时再捡到缩小道具会刷新持续时间。
func (p *Player) Shrink(ticks int) {
	p.SetScale(ItemShrinkScale)
	p.shrinkTicks = ticks
}

// SetScale 设置缩放比例，不会小于 PlayerMinScale，
// 否则角色会小到几乎撞不到管子，游戏就没难度了。
//
// 缩放时保持角色中心不动，否则体型突然变化会让角色看起来「跳」了一下。
func (p *Player) SetScale(scale float64) {
	newScale := math.Max(scale, PlayerMinScale)
	cx := p.X + p.W*p.Scale/2
	cy := p.Y + p.H*p.Scale/2
	p.Scale = newScale
	p.X = cx - p.W*p.Scale/2
	p.Y = cy - p.H*p.Scale/2
}

// Jump 给角色一个向上的初速度。
func (p *Player) Jump() {
	p.VelY = JumpVelocity
}

// Update 按固定步长推进一个 tick 的物理模拟。
func (p *Player) Update() {
	p.VelY += Gravity
	if p.VelY > MaxFallSpeed {
		p.VelY = MaxFallSpeed
	}
	p.Y += p.VelY

	// 缩小是有时限的：时间一到就恢复原尺寸。
	if p.shrinkTicks > 0 {
		p.shrinkTicks--
		if p.shrinkTicks == 0 {
			p.SetScale(1)
		}
	}
}

// Rect 返回当前的碰撞盒（体型会随 Scale 变化）。
func (p *Player) Rect() Rect {
	return Rect{X: p.X, Y: p.Y, W: p.W * p.Scale, H: p.H * p.Scale}
}

// Draw 把角色贴图按当前缩放画到当前位置。
// alpha 是不透明度（0~1），受伤后的无敌时间用它来做闪烁提示。
func (p *Player) Draw(dst, img *ebiten.Image, alpha float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(p.Scale, p.Scale)
	op.GeoM.Translate(p.X, p.Y)
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(img, op)
}

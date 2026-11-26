package game

import "github.com/hajimehoshi/ebiten/v2"

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
	X, Y float64
	W, H float64
	VelY float64 // 竖直速度，单位：像素/tick，向下为正
}

// NewPlayer 在屏幕左侧、垂直居中的位置创建一个静止的角色。
// 尺寸取自贴图本身（gopher.png 是 60x75），这样换素材时不需要改代码。
func NewPlayer(w, h float64) *Player {
	return &Player{
		X: PlayerX,
		Y: (ScreenHeight - h) / 2,
		W: w,
		H: h,
	}
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
}

// Rect 返回当前的碰撞盒。
func (p *Player) Rect() Rect {
	return Rect{X: p.X, Y: p.Y, W: p.W, H: p.H}
}

// Draw 把角色贴图画到当前位置。
func (p *Player) Draw(dst, img *ebiten.Image) {
	drawSprite(dst, img, p.X, p.Y)
}

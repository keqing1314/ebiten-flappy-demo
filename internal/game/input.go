package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input 把多种操作方式（键盘、鼠标、触摸）统一成一个语义化的事件。
//
// 对比 miniten 版本：那里需要自己用 isPrevClicked 记录上一帧的状态，
// 再和新状态比较来判断「刚刚按下」。Ebitengine 的 inpututil 已经内置了这种边沿检测：
//
//	IsKeyJustPressed / IsMouseButtonJustPressed / AppendJustPressedTouchIDs
//
// 它们保证同一次按下只会在一个 tick 里返回 true，因此不需要我们再维护任何历史状态。
type Input struct {
	jumpPressed bool
	menuDelta   int // 本 tick 的菜单左右切换量：-1 向左，+1 向右

	// 本 tick 点击（鼠标或触摸）的位置，用于判断点中了哪个按钮。
	pressX, pressY float64
	hasPressPos    bool

	touchIDs []ebiten.TouchID
}

// Update 读取本 tick 的输入状态，必须在每个 tick 的最开始调用一次。
//
// 因为 Update 的调用频率由 TPS 固定，所以「一次按下 = 一次跳跃」不会被高刷新率放大。
func (in *Input) Update() {
	// 复用同一个切片，避免每 tick 都为触摸 ID 分配内存。
	in.touchIDs = inpututil.AppendJustPressedTouchIDs(in.touchIDs[:0])
	in.jumpPressed = inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		len(in.touchIDs) > 0

	in.menuDelta = 0
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		in.menuDelta--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		in.menuDelta++
	}

	// 记录本 tick 点击的位置：鼠标优先，其次是第一个刚按下的触摸点。
	in.hasPressPos = false
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		in.pressX, in.pressY, in.hasPressPos = float64(x), float64(y), true
	}
	for _, id := range in.touchIDs {
		x, y := ebiten.TouchPosition(id)
		in.pressX, in.pressY, in.hasPressPos = float64(x), float64(y), true
		break
	}
}

// JumpJustPressed 报告本 tick 是否发生了一次新的「拍翅膀」输入。
func (in *Input) JumpJustPressed() bool {
	return in.jumpPressed
}

// MenuDelta 返回本 tick 菜单选择的变化：-1 向左，+1 向右，0 不变。
func (in *Input) MenuDelta() int {
	return in.menuDelta
}

// PressPosition 返回本 tick 点击的位置（逻辑坐标），鼠标和触摸都支持。
// ok 为 false 表示本 tick 的「按下」来自键盘，没有位置信息。
func (in *Input) PressPosition() (x, y float64, ok bool) {
	return in.pressX, in.pressY, in.hasPressPos
}

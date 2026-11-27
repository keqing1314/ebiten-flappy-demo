package game

import "github.com/hajimehoshi/ebiten/v2"

// DisplayMode 是窗口的显示模式：普通窗口，或铺满显示器的全屏。
//
// 全屏由 Ebitengine 负责把 640x360 的逻辑画布等比放大到整个显示器，
// 因此游戏代码仍然只需要按逻辑分辨率绘制，不需要关心真实的屏幕尺寸。
type DisplayMode int

const (
	DisplayWindowed DisplayMode = iota
	DisplayFullscreen

	// numDisplayModes 是显示模式总数，界面按它生成按钮。
	numDisplayModes = int(DisplayFullscreen) + 1
)

// Name 返回界面上显示的名字（内置像素字体只有 ASCII 字形）。
func (m DisplayMode) Name() string {
	if m == DisplayFullscreen {
		return "FULLSCREEN"
	}
	return "WINDOWED"
}

// Display 返回当前的显示模式。
func (g *Game) Display() DisplayMode {
	return g.display
}

// SetDisplay 记录要使用的显示模式。
//
// 真正的窗口切换由 applyDisplay 在本次 Update 结束时统一完成：
// 场景代码只改状态，不直接调用 ebiten 的窗口 API，这样纯逻辑测试也不需要图形环境。
func (g *Game) SetDisplay(m DisplayMode) {
	if m < 0 || int(m) >= numDisplayModes {
		return
	}
	g.display = m
}

// toggleDisplay 在窗口与全屏之间切换。
func (g *Game) toggleDisplay() {
	if g.display == DisplayFullscreen {
		g.SetDisplay(DisplayWindowed)
	} else {
		g.SetDisplay(DisplayFullscreen)
	}
}

// applyDisplay 把设置同步到窗口。
// ebiten.SetFullscreen 只在设置变化时调用一次，避免每个 tick 都重复设置窗口模式。
func (g *Game) applyDisplay() {
	if g.display == g.appliedDisplay {
		return
	}
	ebiten.SetFullscreen(g.display == DisplayFullscreen)
	g.appliedDisplay = g.display
}

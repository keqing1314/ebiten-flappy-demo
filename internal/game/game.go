package game

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/keqing1314/ebiten-flappy-demo/internal/assets"
	"github.com/keqing1314/ebiten-flappy-demo/internal/sound"
)

// Game 实现了 ebiten.Game 接口，这是引擎与我们的代码之间唯一的约定。
// 引擎的游戏循环本质上就是不断重复：
//
//	读取输入 → 调用 Update（按 TPS） → 调用 Draw（按显示器刷新率）
//
// 三个方法的职责：
//   - Update：只做逻辑计算（移动、碰撞、状态切换），不碰任何绘制 API。
//   - Draw：只做绘制，不修改任何会影响下一帧的状态。
//   - Layout：回答「逻辑画布有多大」，缩放与黑边由引擎负责。
//
// 逻辑与绘制必须分开，是游戏循环最重要的习惯：
// Update 的调用次数由 TPS 决定，Draw 的调用次数由显示器刷新率（FPS）决定，两者并不相等。
// 如果在 Draw 里修改状态，游戏速度就会随显示器刷新率变化——这就是很多老游戏
// 在 144Hz 显示器上「跑得特别快」的原因（也就是常说的画面撕裂/垂直同步话题）。
type Game struct {
	assets    *assets.Assets
	fontSmall *text.GoTextFace // 16px，提示文字
	fontLarge *text.GoTextFace // 32px，标题与分数
	sfx       *sound.Sound

	scene Scene
	input Input

	difficulty Difficulty // 当前难度，由标题画面选择，重开一局时沿用

	display        DisplayMode // 当前显示模式：窗口或全屏
	appliedDisplay DisplayMode // 已经应用到窗口上的显示模式

	best  int  // 本次运行内的最高分
	debug bool // 是否显示调试信息（F1 切换）
}

// New 加载资源并创建游戏。返回错误表示资源缺失等致命问题。
func New() (*Game, error) {
	a, err := assets.Load()
	if err != nil {
		return nil, err
	}

	g := &Game{
		assets:    a,
		fontSmall: a.NewFace(16),
		fontLarge: a.NewFace(32),
		sfx:       sound.Load(),
	}
	g.scene = newTitleScene()
	return g, nil
}

// Update 每个 tick 被调用一次。
// 返回 ebiten.Termination 会让 RunGame 立刻返回 nil（正常退出）。
func (g *Game) Update() error {
	g.input.Update()

	// Esc 打开暂停菜单；在暂停菜单里再按一次 Esc 则继续游戏。
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.togglePause()
	}
	// F1 切换调试信息（FPS/TPS、碰撞盒）。
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.debug = !g.debug
	}
	// F / F11 随时切换窗口与全屏，标题画面也可以直接点击对应的按钮。
	if g.input.ToggleDisplayJustPressed() {
		g.toggleDisplay()
	}

	// 具体的游戏逻辑交给当前场景。
	if err := g.scene.Update(g); err != nil {
		return err
	}

	// 场景可能刚刚修改了显示设置，在本 tick 结束时统一应用到窗口。
	g.applyDisplay()
	return nil
}

// Draw 每一帧画面被调用一次。
func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(g, screen)

	if g.debug {
		// DebugPrintAt 用引擎内置的位图字体打印调试信息，不需要加载任何字体资源。
		// ActualTPS/ActualFPS 分别是逻辑更新与画面绘制的实际频率。
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("TPS %.1f  FPS %.1f", ebiten.ActualTPS(), ebiten.ActualFPS()),
			4, ScreenHeight-16)
	}
}

// Layout 返回逻辑画布的尺寸。
//
// 窗口被拖动缩放时，引擎会把 640x360 的画面等比放大/缩小，多出来的区域补黑边。
// 因此我们永远只需要按 640x360 来写绘制代码，不用关心真实窗口有多少像素。
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// switchTo 切换当前场景。
// 切换只在 Update 里进行，避免在一帧绘制的过程中改变绘制内容。
func (g *Game) switchTo(s Scene) {
	g.scene = s
}

// recordBest 更新本次运行内的最高分。
func (g *Game) recordBest(score int) {
	if score > g.best {
		g.best = score
	}
}

// Icon 返回窗口图标，供 main 在创建窗口时使用。
func (g *Game) Icon() image.Image {
	return g.assets.Icon()
}

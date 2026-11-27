package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// 暂停菜单的布局。
const (
	pauseButtonW   = 220.0
	pauseButtonH   = 44.0
	pauseButtonGap = 32.0
	pauseButtonY   = 160.0
)

// pauseOption 是暂停菜单里的选项。
type pauseOption int

const (
	pauseOptionMainMenu pauseOption = iota
	pauseOptionQuit

	// numPauseOptions 是选项总数，界面按它生成按钮、做循环选择。
	numPauseOptions = int(pauseOptionQuit) + 1
)

// Name 返回界面上显示的名字（内置像素字体只有 ASCII 字形）。
func (o pauseOption) Name() string {
	if o == pauseOptionQuit {
		return "QUIT"
	}
	return "MAIN MENU"
}

// pauseButtonRect 返回第 i 个暂停菜单按钮的矩形。
func pauseButtonRect(i int) Rect {
	return centeredButtonRect(i, numPauseOptions, pauseButtonY, pauseButtonW, pauseButtonH, pauseButtonGap)
}

// hitTestPause 返回点 (x, y) 命中的暂停菜单按钮，没点中任何按钮时 ok 为 false。
func hitTestPause(x, y float64) (pauseOption, bool) {
	for i := 0; i < numPauseOptions; i++ {
		if pauseButtonRect(i).Contains(x, y) {
			return pauseOption(i), true
		}
	}
	return 0, false
}

// pauseScene 是暂停菜单：盖在原来的画面上，让玩家选择返回主界面或退出游戏。
//
// 它不推进原来的场景，所以背景画面会定格；按 Esc 关闭菜单即可继续。
type pauseScene struct {
	prev     Scene
	selected pauseOption
}

func newPauseScene(prev Scene) *pauseScene {
	return &pauseScene{prev: prev}
}

// togglePause 在暂停菜单与原来的画面之间切换（Esc 键）。
// 暂停菜单里再按一次 Esc 就恢复到进入菜单前的场景。
func (g *Game) togglePause() {
	if p, ok := g.scene.(*pauseScene); ok {
		g.switchTo(p.prev)
		return
	}
	g.switchTo(newPauseScene(g.scene))
}

// confirm 返回本 tick 是否确认了某个选项，以及确认的是哪一个：
//   - 键盘 Space/Enter/↑/W：确认当前高亮的选项；
//   - 鼠标/触摸：点中按钮才确认，点空白处什么也不做。
func (s *pauseScene) confirm(in Input) (pauseOption, bool) {
	if x, y, ok := in.PressPosition(); ok {
		return hitTestPause(x, y)
	}
	if in.JumpJustPressed() {
		return s.selected, true
	}
	return 0, false
}

func (s *pauseScene) Update(g *Game) error {
	// 左右（或 A/D）切换选项，到头后循环。
	if d := g.input.MenuDelta(); d != 0 {
		s.selected = pauseOption((int(s.selected) + d + numPauseOptions) % numPauseOptions)
	}

	choice, ok := s.confirm(g.input)
	if !ok {
		return nil
	}

	if choice == pauseOptionQuit {
		return ebiten.Termination
	}

	g.switchTo(newTitleSceneWithDifficulty(g.difficulty))
	return nil
}

func (s *pauseScene) Draw(g *Game, screen *ebiten.Image) {
	// 先画暂停前的画面，再压暗，最后叠上菜单。
	s.prev.Draw(g, screen)
	fillOverlay(screen, 170)

	drawTextCentered(screen, "PAUSED", g.fontLarge, 80, color.White)

	for i := 0; i < numPauseOptions; i++ {
		drawOptionButton(screen, pauseButtonRect(i), pauseOption(i).Name(), g.fontSmall, pauseOption(i) == s.selected)
	}

	drawTextCentered(screen, "LEFT/RIGHT SELECT  SPACE CONFIRM", g.fontSmall, 240, color.White)
	drawTextCentered(screen, "ESC RESUME", g.fontSmall, 262, color.White)
}

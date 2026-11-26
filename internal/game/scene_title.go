package game

import (
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// 标题画面上难度与显示模式按钮的布局。
const (
	difficultyButtonW   = 150.0
	difficultyButtonH   = 44.0
	difficultyButtonGap = 32.0
	difficultyButtonY   = 180.0

	displayButtonW   = 170.0
	displayButtonH   = 30.0
	displayButtonGap = 32.0
	displayButtonY   = 232.0
)

// titleScene 是标题画面：选择难度、显示模式并开始游戏。
type titleScene struct {
	ticks    int
	selected Difficulty
}

func newTitleScene() *titleScene {
	return &titleScene{selected: DifficultyNormal}
}

// difficultyButtonRect 返回第 i 个难度按钮的矩形。
func difficultyButtonRect(i int) Rect {
	total := float64(numDifficulties)*difficultyButtonW + float64(numDifficulties-1)*difficultyButtonGap
	x := (ScreenWidth - total) / 2
	return Rect{
		X: x + float64(i)*(difficultyButtonW+difficultyButtonGap),
		Y: difficultyButtonY,
		W: difficultyButtonW,
		H: difficultyButtonH,
	}
}

// hitTestDifficulty 返回点 (x, y) 命中的难度按钮，没点中任何按钮时 ok 为 false。
func hitTestDifficulty(x, y float64) (Difficulty, bool) {
	for i := 0; i < numDifficulties; i++ {
		if difficultyButtonRect(i).Contains(x, y) {
			return Difficulty(i), true
		}
	}
	return 0, false
}

// displayButtonRect 返回第 i 个显示模式按钮的矩形。
func displayButtonRect(i int) Rect {
	total := float64(numDisplayModes)*displayButtonW + float64(numDisplayModes-1)*displayButtonGap
	x := (ScreenWidth - total) / 2
	return Rect{
		X: x + float64(i)*(displayButtonW+displayButtonGap),
		Y: displayButtonY,
		W: displayButtonW,
		H: displayButtonH,
	}
}

// hitTestDisplay 返回点 (x, y) 命中的显示模式按钮，没点中任何按钮时 ok 为 false。
func hitTestDisplay(x, y float64) (DisplayMode, bool) {
	for i := 0; i < numDisplayModes; i++ {
		if displayButtonRect(i).Contains(x, y) {
			return DisplayMode(i), true
		}
	}
	return 0, false
}

// displayChoice 返回本 tick 点击命中的显示模式按钮。
// 点击设置按钮只切换设置，不应该顺带开始游戏。
func displayChoice(in Input) (DisplayMode, bool) {
	if x, y, ok := in.PressPosition(); ok {
		return hitTestDisplay(x, y)
	}
	return 0, false
}

// startDifficulty 返回按下「开始」时应当使用的难度：
// 用鼠标/触摸点中按钮就用那个难度，否则沿用当前选中的难度。
func (s *titleScene) startDifficulty(in Input) Difficulty {
	if x, y, ok := in.PressPosition(); ok {
		if d, hit := hitTestDifficulty(x, y); hit {
			return d
		}
	}
	return s.selected
}

// Update 处理难度与显示模式选择，并在玩家按下「开始」时进入游戏。
func (s *titleScene) Update(g *Game) error {
	s.ticks++

	// 左右（或 A/D）切换难度，到头后循环。
	if d := g.input.MenuDelta(); d != 0 {
		s.selected = Difficulty((int(s.selected) + d + numDifficulties) % numDifficulties)
	}

	if !g.input.JumpJustPressed() {
		return nil
	}

	// 点中显示模式按钮时只切换设置，不开始游戏。
	if m, hit := displayChoice(g.input); hit {
		g.SetDisplay(m)
		return nil
	}

	s.selected = s.startDifficulty(g.input)
	g.difficulty = s.selected
	g.switchTo(newPlayScene(g))
	return nil
}

func (s *titleScene) Draw(g *Game, screen *ebiten.Image) {
	drawBackground(screen, g.assets.Sky)

	// 让地鼠轻轻地上下浮动，静止画面也显得有生气。
	bob := math.Sin(float64(s.ticks)/15) * 8
	size := g.assets.Gopher.Bounds()
	drawSprite(screen, g.assets.Gopher, (ScreenWidth-float64(size.Dx()))/2, 85+bob)

	drawTextCentered(screen, "FLAPPY GOPHER", g.fontLarge, 26, color.White)

	// 难度按钮：选中的高亮，未选中的压暗。
	for i := 0; i < numDifficulties; i++ {
		s.drawDifficultyButton(g, screen, Difficulty(i))
	}

	// 显示模式按钮：当前模式高亮。
	for i := 0; i < numDisplayModes; i++ {
		s.drawDisplayButton(g, screen, DisplayMode(i))
	}

	if g.best > 0 {
		drawTextCentered(screen, "BEST "+strconv.Itoa(g.best), g.fontSmall, 276, color.White)
	}

	drawTextCentered(screen, "F TOGGLE WINDOW / FULLSCREEN", g.fontSmall, 300, color.White)

	// 提示文字闪烁：每 30 个 tick（约半秒）切换一次显示状态。
	if (s.ticks/30)%2 == 0 {
		drawTextCentered(screen, "ARROWS TO SELECT  SPACE TO START", g.fontSmall, 322, color.White)
	}
}

// drawDifficultyButton 绘制一个难度按钮。
func (s *titleScene) drawDifficultyButton(g *Game, screen *ebiten.Image, d Difficulty) {
	drawOptionButton(screen, difficultyButtonRect(int(d)), d.Name(), g.fontSmall, d == s.selected)
}

// drawDisplayButton 绘制一个显示模式按钮。
func (s *titleScene) drawDisplayButton(g *Game, screen *ebiten.Image, m DisplayMode) {
	drawOptionButton(screen, displayButtonRect(int(m)), m.Name(), g.fontSmall, m == g.Display())
}

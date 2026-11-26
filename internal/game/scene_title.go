package game

import (
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// 标题画面上难度按钮的布局。
const (
	difficultyButtonW      = 150.0
	difficultyButtonH      = 44.0
	difficultyButtonGap    = 32.0
	difficultyButtonY      = 180.0
	difficultyButtonBorder = 2.0
)

// titleScene 是标题画面：选择难度并开始游戏。
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

// Update 处理难度选择，并在玩家按下「开始」时进入游戏。
func (s *titleScene) Update(g *Game) error {
	s.ticks++

	// 左右（或 A/D）切换难度，到头后循环。
	if d := g.input.MenuDelta(); d != 0 {
		s.selected = Difficulty((int(s.selected) + d + numDifficulties) % numDifficulties)
	}

	if !g.input.JumpJustPressed() {
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

	if g.best > 0 {
		drawTextCentered(screen, "BEST "+strconv.Itoa(g.best), g.fontSmall, 240, color.White)
	}

	// 提示文字闪烁：每 30 个 tick（约半秒）切换一次显示状态。
	if (s.ticks/30)%2 == 0 {
		drawTextCentered(screen, "ARROWS TO SELECT  SPACE TO START", g.fontSmall, 290, color.White)
	}
}

// drawDifficultyButton 绘制一个难度按钮。
func (s *titleScene) drawDifficultyButton(g *Game, screen *ebiten.Image, d Difficulty) {
	r := difficultyButtonRect(int(d))
	selected := d == s.selected

	// 选中项用白底深色字（反色高亮），未选中项压暗，保证两者都清晰可读。
	var fill, border, textColor color.Color = color.RGBA{A: 0x50},
		color.RGBA{R: 0x90, G: 0x90, B: 0x90, A: 0xff},
		color.RGBA{R: 0xa0, G: 0xa0, B: 0xa0, A: 0xff}
	if selected {
		fill = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		border = color.White
		textColor = color.RGBA{R: 0x20, G: 0x40, B: 0x60, A: 0xff}
	}

	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), fill, false)
	vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), difficultyButtonBorder, border, false)
	drawTextCenteredIn(screen, d.Name(), g.fontSmall, r, textColor)
}

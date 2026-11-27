package game

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// restartDelay 是失败后多久才开始接受菜单输入（单位：tick）。
// 玩家失误时往往还按着键，加一点延迟可以避免「还没看清分数就又开局了」。
const restartDelay = 20

// 失败菜单的布局。
const (
	gameOverButtonW   = 180.0
	gameOverButtonH   = 44.0
	gameOverButtonGap = 32.0
	gameOverButtonY   = 232.0
)

// gameOverOption 是失败画面上的选项。
type gameOverOption int

const (
	gameOverRestart gameOverOption = iota
	gameOverQuit

	// numGameOverOptions 是选项总数，界面按它生成按钮、做循环选择。
	numGameOverOptions = int(gameOverQuit) + 1
)

// Name 返回界面上显示的名字（内置像素字体只有 ASCII 字形）。
func (o gameOverOption) Name() string {
	if o == gameOverQuit {
		return "QUIT"
	}
	return "RESTART"
}

// gameOverButtonRect 返回第 i 个失败菜单按钮的矩形。
func gameOverButtonRect(i int) Rect {
	return centeredButtonRect(i, numGameOverOptions, gameOverButtonY, gameOverButtonW, gameOverButtonH, gameOverButtonGap)
}

// hitTestGameOver 返回点 (x, y) 命中的失败菜单按钮，没点中任何按钮时 ok 为 false。
func hitTestGameOver(x, y float64) (gameOverOption, bool) {
	for i := 0; i < numGameOverOptions; i++ {
		if gameOverButtonRect(i).Contains(x, y) {
			return gameOverOption(i), true
		}
	}
	return 0, false
}

// gameOverScene 是游戏结束画面。
// 它保留了对失败那一刻 playScene 的引用，用来把游戏世界「定格」在背景上，
// 让玩家看清楚自己是撞上了哪根管子。
type gameOverScene struct {
	prev     *playScene
	ticks    int
	selected gameOverOption
}

func newGameOverScene(prev *playScene) *gameOverScene {
	return &gameOverScene{prev: prev, selected: gameOverRestart}
}

// confirm 返回本 tick 是否确认了某个选项，以及确认的是哪一个：
//   - 键盘 Space/Enter/↑/W：确认当前高亮的选项；
//   - 鼠标/触摸：点中按钮才确认，点空白处什么也不做。
func (s *gameOverScene) confirm(in Input) (gameOverOption, bool) {
	if x, y, ok := in.PressPosition(); ok {
		return hitTestGameOver(x, y)
	}
	if in.JumpJustPressed() {
		return s.selected, true
	}
	return 0, false
}

func (s *gameOverScene) Update(g *Game) error {
	s.ticks++
	if s.ticks < restartDelay {
		return nil
	}

	// 左右（或 A/D）切换选项，到头后循环。
	if d := g.input.MenuDelta(); d != 0 {
		s.selected = gameOverOption((int(s.selected) + d + numGameOverOptions) % numGameOverOptions)
	}

	choice, ok := s.confirm(g.input)
	if !ok {
		return nil
	}

	if choice == gameOverQuit {
		return ebiten.Termination
	}

	g.switchTo(newPlayScene(g))
	return nil
}

func (s *gameOverScene) Draw(g *Game, screen *ebiten.Image) {
	// 先画定格的世界，再压暗，最后叠上文字与菜单。
	s.prev.drawWorld(g, screen)
	fillOverlay(screen, 150)

	score := s.prev.pipes.Score()

	drawTextCentered(screen, "GAME OVER", g.fontLarge, 80, color.White)
	drawTextCentered(screen, "SCORE "+strconv.Itoa(score), g.fontSmall, 140, color.White)
	drawTextCentered(screen, "BEST "+strconv.Itoa(g.best), g.fontSmall, 170, color.White)

	// 刷新纪录时给一点正反馈。
	if score > 0 && score == g.best {
		drawTextCentered(screen, "NEW BEST!", g.fontSmall, 200, color.RGBA{R: 0xff, G: 0xd7, A: 0xff})
	}

	for i := 0; i < numGameOverOptions; i++ {
		drawOptionButton(screen, gameOverButtonRect(i), gameOverOption(i).Name(), g.fontSmall, gameOverOption(i) == s.selected)
	}

	drawTextCentered(screen, "LEFT/RIGHT SELECT  SPACE CONFIRM", g.fontSmall, 292, color.White)
}

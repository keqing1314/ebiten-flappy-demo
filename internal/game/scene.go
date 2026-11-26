package game

import "github.com/hajimehoshi/ebiten/v2"

// Scene 表示游戏的一个画面：标题、游戏中、游戏结束。
//
// Game 只持有「当前场景」一个引用，把 Update/Draw 全部转发出去。
// 这样每种画面都是互不干扰的一小块代码，
// 比在一个巨大的 switch 里堆所有逻辑更容易阅读和扩展
// （想加一个「排行榜」画面，只需要再写一个实现了 Scene 的类型）。
type Scene interface {
	// Update 推进一个 tick。只在 Update 里改变游戏状态。
	Update(g *Game) error

	// Draw 绘制一帧。只在 Draw 里调用绘制 API。
	Draw(g *Game, screen *ebiten.Image)
}

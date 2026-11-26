// Flappy Gopher 是一个用完整 Ebitengine (v2) API 编写的入门示例。
//
// 与最初的 miniten 版本相比，这里不再使用任何「简化封装」：
//   - 窗口与帧循环：自己实现 ebiten.Game 接口（Update / Draw / Layout）
//   - 输入：用 inpututil 做「本 tick 刚刚按下」的边沿检测
//   - 文字：用 text/v2 渲染矢量字体（支持任意字号）
//   - 音效：用 audio 包播放现场合成出来的 PCM 数据
//   - 资源：用 go:embed 打包进二进制，单个 exe 即可运行
//
// 代码结构：
//
//	main.go              程序入口：配置窗口并启动游戏循环
//	internal/game/       游戏逻辑与绘制（场景、角色、管子、输入、工具函数）
//	internal/assets/     贴图与字体（含 go:embed）
//	internal/sound/      代码合成的音效
package main

import (
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/keqing1314/ebiten-flappy-demo/internal/game"
)

func main() {
	g, err := game.New()
	if err != nil {
		log.Fatalf("初始化游戏失败: %v", err)
	}

	// ---- 窗口设置，必须在 RunGame 之前调用 ----
	ebiten.SetWindowTitle("Flappy Gopher - Ebitengine Demo")
	ebiten.SetWindowIcon([]image.Image{g.Icon()})

	// 逻辑分辨率固定 640x360，窗口默认放大 2 倍成 1280x720。
	ebiten.SetWindowSize(game.ScreenWidth*game.WindowScale, game.ScreenHeight*game.WindowScale)

	// 允许拖动窗口自由缩放。因为 Layout 固定返回 640x360，
	// 引擎会把画面等比缩放并在多出来的地方补黑边，画面不会变形。
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// TPS（Ticks Per Second）是 Update 每秒被调用的次数，与显示器刷新率（FPS）是两回事。
	// 固定 60 表示逻辑稳定地以每秒 60 步推进，与显示器是 60Hz 还是 144Hz 无关。
	// 想体验「帧率越高跑得越快」的经典 bug，把物理参数挪到 Draw 里就能复现 :)
	ebiten.SetTPS(game.TPS)

	// RunGame 会一直阻塞到窗口关闭。它内部就是一个游戏循环：
	// 读取输入 → 按 TPS 调用 Game.Update → 按 FPS 调用 Game.Draw → 交换缓冲区。
	// 返回 ebiten.Termination（例如在 Update 里按了 Esc）时，RunGame 返回 nil。
	if err := ebiten.RunGame(g); err != nil {
		log.Fatalf("游戏异常退出: %v", err)
	}
}

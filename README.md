# Flappy Gopher

一个用 **完整 Ebitengine v2**（`github.com/hajimehoshi/ebiten/v2`）重写的入门示例。

最初的版本是基于 [miniten](https://github.com/eihigh/miniten) 的教学 demo——miniten 把窗口、帧循环、
输入、文字都封装好了，适合第一天上手，但看不到引擎真正的样子。
这个版本不再使用任何封装，全部改用 Ebitengine 原生 API，并按真实项目的结构拆分文件。

## 运行

```bash
go run .
```

编译成单个可执行文件（资源已通过 `go:embed` 打包进去，exe 可独立分发）：

```bash
go build -o ebiten-flappy-demo.exe .
```

环境要求：Go 1.25+。Windows / macOS / Linux 均可直接运行，无需安装任何 SDK 或运行库。

## 操作

| 操作 | 按键 |
| --- | --- |
| 拍翅膀 / 开始 / 重来 | 鼠标左键、触摸、`Space`、`Enter`、`↑`、`W` |
| 显示调试信息（FPS/TPS、碰撞盒） | `F1` |
| 退出 | `Esc` |

> 界面文字使用英文：内置的像素字体 PressStart2P 只包含 ASCII 字形。
> 想显示中文，需要在 `internal/assets` 里换成支持中文的字体（例如 M+ 或思源黑体）。

## 血量机制

角色默认有 **6 颗心**，左上角的血条会实时显示剩余血量（已失去的心用半透明图标占位）。

| 碰撞对象 | 效果 |
| --- | --- |
| 管子 | 扣 1 颗心，播放受伤音效 |
| 天花板 | 扣 1 颗心，角色被贴回屏幕内 |
| 地板 | 扣 1 颗心，角色被贴回屏幕内 |

- 每次受伤后有 60 tick（1 秒）的**无敌时间**，期间角色半透明闪烁，再次碰撞不会扣血。
  没有它的话，角色贴住管子/地板的那几十个 tick 会被连续判定为碰撞，一次失误就掉光所有心。
- 有头盔护盾时，碰撞优先消耗护盾，不扣心血（见下面的「道具」）。
- 血量扣到 0 才算失败：此时播放死亡音效并进入结束画面（最后一击只播死亡音效，不叠受伤音效）。
- 血量、护盾、无敌时间的逻辑集中在 `internal/game/health.go`，与绘制、音效完全解耦，可以直接写单元测试。

## 道具

道具随机出现在管子空隙的中心，跟着管子一起移动，角色经过时自动拾取。
每根新管子有 30% 的概率携带一个道具，种类按权重随机：

| 道具 | 效果 | 出现条件 | 权重 |
| --- | --- | --- | --- |
| 心 | 恢复 1 颗心（不超过上限） | 血不满时 | 40 |
| 头盔 | 获得 3 层护盾，碰撞只消耗护盾、不扣血 | 护盾未满时 | 30 |
| 缩小 | 角色缩小到 1/3，碰撞体积同步变小 | 尚未缩小时 | 20 |
| 减速 | 永久降低卷动速度 0.4 像素/tick（不低于 1.8） | 当前速度 ≥ 3.0 时 | 10 |

- 头盔图标画在血条后面，每被撞一次就变淡一层，用完消失。
- 缩小有下限（`PlayerMinScale`），不会小到几乎撞不到东西；缩放时角色中心保持不动。
- 减速只在速度够快时才会刷出，并且不会把速度压到 `ScrollSpeedFloor` 以下，避免前期过早变简单。
- 拾取先于碰撞判定：刚捡到的头盔能挡住同一 tick 的碰撞，缩小也能让本 tick 的碰撞盒立刻变小。

**扩展新道具**只需要三步：在 `item.go` 加一个 `ItemKind` 常量、把贴图放进 `internal/assets/images`
并在 `assets.go` 里加载、在 `itemSpecs` 表里加一项（权重、出现条件、贴图、`Apply` 效果）。
生成、绘制、拾取流程（`PipeField` / `playScene`）都不需要改动。

## 项目结构

```text
ebiten-flappy-demo/
├── main.go                      程序入口：窗口设置 + 启动游戏循环
├── go.mod / go.sum
└── internal/
    ├── assets/
    │   ├── assets.go            go:embed 加载贴图与字体
    │   └── images/              sky / wall / gopher / heart 与四个道具图标
    ├── game/
    │   ├── game.go              Game：实现 ebiten.Game（Update / Draw / Layout）
    │   ├── scene.go             Scene 接口（画面状态机）
    │   ├── scene_title.go       标题画面
    │   ├── scene_play.go        游戏画面
    │   ├── scene_gameover.go    结束画面
    │   ├── player.go            主角：重力 + 跳跃的速度积分
    │   ├── health.go            血量、护盾与受伤后的无敌时间
    │   ├── item.go              道具系统（种类登记表 / 出现条件 / 拾取效果）
    │   ├── pipe.go              管子与管子管理器（生成 / 移动 / 计分 / 回收 / 携带道具）
    │   ├── input.go             统一的输入抽象（键盘 / 鼠标 / 触摸）
    │   ├── draw.go              绘制工具（贴图、带阴影的文字、调试线框）
    │   ├── rect.go              矩形与 AABB 碰撞
    │   ├── config.go            所有可调参数
    │   └── game_test.go         纯逻辑单元测试
    └── sound/
        ├── sound.go             合成音效 + 解码内嵌 mp3
        ├── hurt.mp3             受伤音效素材（go:embed 打包）
        └── death.mp3            死亡音效素材（go:embed 打包）
```

## 和 miniten 版本的对应关系

| miniten | Ebitengine |
| --- | --- |
| `miniten.Run(draw)` | 实现 `ebiten.Game` 接口 + `ebiten.RunGame(g)` |
| 全局变量 + `draw()` 里既算又画 | `Update()` 只算逻辑，`Draw()` 只画画面 |
| `miniten.IsClicked()` + 手写 `isPrevClicked` | `inpututil.IsKeyJustPressed` / `IsMouseButtonJustPressed` |
| `miniten.Println(...)` | `text/v2` 渲染真实字体（任意字号、对齐、描边） |
| `miniten.DrawImage("x.png", x, y)` | `ebitenutil` 载入为 `*ebiten.Image` + `DrawImage` + `GeoM` |
| 固定 640x360、无法缩放 | `Layout` 定义逻辑分辨率，窗口可自由拖动缩放，引擎负责等比缩放与黑边 |

## 关键概念（对应到代码）

**游戏循环与 TPS/FPS**
引擎反复调用 `Game.Update`（固定 60 次/秒，由 `ebiten.SetTPS` 决定）和 `Game.Draw`
（由显示器刷新率决定）。两者频率不同，所以**绝不能在 `Draw` 里改状态**，否则游戏速度会随显示器变化。
调试模式下底部会显示实际的 TPS/FPS（见 `game.go`）。

**逻辑分辨率**
`Layout` 固定返回 `640x360`，绘制代码只关心这块虚拟画布，窗口大小、缩放、黑边全部交给引擎。

**速度积分**
`player.go` 用的是 `速度 += 重力；位置 += 速度` 的固定步长积分，
配合 `MaxFallSpeed` 防止高速下落时穿透障碍物。

**碰撞检测**
`rect.go` 实现轴对齐矩形（AABB）相交判定，`pipe.go` 把每根管子拆成上下两个矩形。
按 `F1` 可以看到实际参与判定的矩形。

**难度曲线**
`pipe.go` 的 `ScrollSpeedAt` 让卷动速度随存活时间按对数增长：
`speed = min(ScrollSpeedMax, ScrollSpeed + ScrollSpeedGain*ln(1+ticks/ScrollSpeedRampTicks))`。
前期提速明显、之后逐渐放缓，最终停在 `ScrollSpeedMax`（约 1.8 倍开局速度）。
对数曲线天然有上限，既能感觉到难度在涨，又不会快到无法反应。

**资源打包**
`assets.go` 用 `//go:embed images/*.png` 把图片编译进二进制。

**音效**
`sound.go` 直接用方波算出 PCM 数据（16 位、双声道、44100Hz），
交给 `audio.NewPlayerFromBytes` 播放。受伤、死亡音效 `hurt.mp3` / `death.mp3`
用 `go:embed` 打包进二进制，运行时由 `audio/mp3` 流式解码，并按音频上下文的采样率重采样。
没有音频设备时会自动静音，不影响游戏。

## 可调参数

都在 [`internal/game/config.go`](internal/game/config.go)：

| 参数 | 含义 | 默认值 |
| --- | --- | --- |
| `ScreenWidth` / `ScreenHeight` | 逻辑分辨率 | 640 x 360 |
| `WindowScale` | 窗口相对逻辑分辨率的倍数 | 2 |
| `TPS` | 每秒逻辑更新次数 | 60 |
| `Gravity` | 重力加速度（像素/tick²） | 0.4 |
| `JumpVelocity` | 每次拍翅膀的初速度 | -6.5 |
| `MaxFallSpeed` | 下落速度上限 | 9.0 |
| `ScrollSpeed` | 开局卷动速度（像素/tick） | 2.2 |
| `ScrollSpeedMax` | 卷动速度上限（存活再久也不会超过） | 4.0 |
| `ScrollSpeedRampTicks` | 提速的时间尺度（tick），越大前期提速越慢 | 1200 |
| `ScrollSpeedGain` | 对数曲线的增速系数 | 0.8 |
| `PipeInterval` | 每隔多少 tick 生成一根管子 | 100 |
| `PipeGapHeight` | 空隙高度 | 150 |
| `MaxGapDelta` | 相邻管子空隙中心的最大高度差 | 80 |
| `MaxHearts` | 初始血量（心数） | 6 |
| `InvulnTicks` | 受伤后的无敌时间（tick） | 60 |
| `HeartSize` | HUD 上心形图标的显示尺寸 | 22 |
| `ItemSpawnChance` | 每根管子携带道具的概率 | 0.3 |
| `ItemSize` | 道具的显示尺寸 / 碰撞盒边长 | 40 |
| `ItemHelmetCharges` / `MaxShield` | 头盔免疫次数 / 护盾上限 | 3 / 3 |
| `ItemShrinkScale` / `PlayerMinScale` | 缩小比例 / 缩小下限 | 1/3 |
| `ItemSlowAmount` | 单次减速量（像素/tick） | 0.4 |
| `ScrollSpeedFloor` | 减速后的速度下限 | 1.8 |
| `ItemSlowMinSpeed` | 减速道具出现的速度门槛 | 3.0 |

## 测试

```bash
go test ./...
```

覆盖的是不依赖窗口/GPU 的纯逻辑：矩形碰撞、跳跃与重力、管子生成节奏、计分与回收、
相邻管子高度差限制，血量机制（无敌时间内不重复扣血、血量耗尽才判负、
撞天花板/地板会扣血并把角色贴回屏幕内），道具（护盾吸收碰撞、回血封顶、缩小下限、
出现条件与拾取），以及随存活时间提升的卷动速度（对数增长、增速逐渐放缓、始终不超过上限）。
这类逻辑抽成纯函数后，测试跑得飞快，也不需要图形环境。

## 可以继续练手的点

- 背景无缝滚动（两张图接力平移，用 `GeoM` 平移）
- 用 `audio.NewInfiniteLoop` 加一段循环背景音乐
- 把最高分写进本地 JSON 文件，实现真正的「一直保存」
- 碰撞盒可视化做成可切换的开关，观察「视觉边缘」和「判定边缘」的差异
- 试试 `ebiten.SetCursorMode`、`ebiten.SetFullscreen`、自定义着色器（`Kage`）

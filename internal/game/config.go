package game

// 本文件集中存放「可调参数」。把数值集中在一处，调手感时只需要改这里。
//
// 坐标系说明：
// Ebitengine 的屏幕坐标原点在【左上角】，x 向右增大，y 向下增大，
// 可以把屏幕当成数学上的第四象限来看——y 越小越靠上。
// 所以「向上跳」在代码里就是给竖直速度一个负值。

const (
	// ScreenWidth、ScreenHeight 是「逻辑分辨率」，也就是游戏内部使用的那块虚拟画布。
	// 绘制代码只需按这个尺寸写，放大/缩小、黑边补全交给引擎的 Layout 处理。
	ScreenWidth  = 640
	ScreenHeight = 360

	// WindowScale 是窗口相对逻辑分辨率的倍数，2 表示默认 1280x720 的窗口。
	WindowScale = 2

	// TPS（Ticks Per Second）是 Update 每秒被调用的次数。
	// 固定 TPS 意味着每个 tick 的时长固定为 1/60 秒，
	// 于是下面所有「每 tick」的物理参数在任何机器上表现都完全一致。
	TPS = 60
)

const (
	// Gravity 是重力加速度，单位：像素/tick²。每个 tick 都把这一值累加到下落速度上。
	Gravity = 0.4
	// JumpVelocity 是拍一次翅膀赋予的竖直速度，负值表示向上。
	JumpVelocity = -6.5
	// MaxFallSpeed 是下落速度上限。没有它的话，速度太快时一帧位移可能超过障碍物宽度，
	// 出现「穿模」（直接从管子里穿过去）。
	MaxFallSpeed = 9.0

	// PlayerX 是角色固定的水平位置（左边缘）。角色只做上下运动，横向不移动。
	PlayerX = 120.0

	// ScrollSpeed 是场景每 tick 向左卷动的距离，单位：像素/tick。
	ScrollSpeed = 2.2
	// PipeWidth 是管子的碰撞宽度，与 wall.png 的宽度（21px）一致。
	PipeWidth = 21.0
	// PipeInterval 是每隔多少个 tick 生成一根新管子。
	PipeInterval = 100
	// PipeGapHeight 是上下两根管子之间的空隙高度，角色要从这里穿过。
	PipeGapHeight = 150.0
	// PipeGapMargin 是空隙中心距离屏幕上/下边缘的最小距离，
	// 避免空隙贴着屏幕边缘生成，玩家完全没有反应时间。
	PipeGapMargin = 25.0
	// MaxGapDelta 限制相邻两根管子空隙中心的高度差。
	// 一次跳跃大约能爬升 50px，如果高度差太大就会出现「根本来不及飞过去」的死局。
	MaxGapDelta = 80.0
)

// 空隙中心的取值范围。空隙的上下边缘必须留在屏幕内（再留出 margin），
// 因此中心点的活动范围是 [上半段, 下半段]。
const (
	minGapCenter = PipeGapHeight/2 + PipeGapMargin
	maxGapCenter = ScreenHeight - PipeGapHeight/2 - PipeGapMargin
)

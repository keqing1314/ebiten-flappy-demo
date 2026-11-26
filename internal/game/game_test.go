package game

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/keqing1314/ebiten-flappy-demo/internal/sound"
)

func TestRectIntersects(t *testing.T) {
	base := Rect{X: 0, Y: 0, W: 10, H: 10}

	tests := []struct {
		name  string
		other Rect
		want  bool
	}{
		{"完全重合", Rect{X: 0, Y: 0, W: 10, H: 10}, true},
		{"部分重叠", Rect{X: 5, Y: 5, W: 10, H: 10}, true},
		{"被完全包含", Rect{X: 2, Y: 2, W: 3, H: 3}, true},
		{"右侧贴合不算碰撞", Rect{X: 10, Y: 0, W: 10, H: 10}, false},
		{"下方贴合不算碰撞", Rect{X: 0, Y: 10, W: 10, H: 10}, false},
		{"完全分离", Rect{X: 100, Y: 100, W: 10, H: 10}, false},
	}

	for _, tt := range tests {
		if got := base.Intersects(tt.other); got != tt.want {
			t.Errorf("%s: Intersects() = %v, 期望 %v", tt.name, got, tt.want)
		}
	}
}

func TestPlayerJumpThenFall(t *testing.T) {
	p := NewPlayer(60, 75)
	if want := float64(ScreenHeight-75) / 2; p.Y != want {
		t.Fatalf("初始 y = %v, 期望 %v", p.Y, want)
	}

	startY := p.Y
	p.Jump()
	if p.VelY != JumpVelocity {
		t.Fatalf("跳跃后 VelY = %v, 期望 %v", p.VelY, JumpVelocity)
	}

	p.Update()
	if p.Y >= startY {
		t.Fatalf("跳跃后应当向上移动, y = %v, 起点 = %v", p.Y, startY)
	}

	// 一直不操作，下落速度会加速到上限并保持不变。
	for i := 0; i < 300; i++ {
		p.Update()
	}
	if p.VelY != MaxFallSpeed {
		t.Fatalf("下落速度应被限制为 %v, 实际 %v", MaxFallSpeed, p.VelY)
	}
	if p.Y <= startY {
		t.Fatalf("长时间不操作后应当落回下方, y = %v", p.Y)
	}
}

func TestPipeFieldSpawn(t *testing.T) {
	f := NewPipeField(42, DifficultyNormal)
	if len(f.Pipes()) != 1 {
		t.Fatalf("开局应当生成 1 根管子, 实际 %d", len(f.Pipes()))
	}

	p := f.Pipes()[0]
	if p.X != ScreenWidth {
		t.Errorf("新管子应当从屏幕右边缘进入, X = %v", p.X)
	}
	if p.GapY < PipeGapMargin {
		t.Errorf("空隙上边缘 %v 越过了上边距 %v", p.GapY, PipeGapMargin)
	}
	if bottom := p.GapY + PipeGapHeight; bottom > ScreenHeight-PipeGapMargin {
		t.Errorf("空隙下边缘 %v 越过了下边距", bottom)
	}
}

func TestPipeFieldSpawnInterval(t *testing.T) {
	f := NewPipeField(1, DifficultyNormal)

	// 关掉计分干扰：这里只关心生成节奏。
	for i := 0; i < PipeInterval-1; i++ {
		f.Update(0)
	}
	if n := len(f.Pipes()); n != 1 {
		t.Fatalf("%d 个 tick 后应有 1 根管子, 实际 %d", PipeInterval-1, n)
	}

	f.Update(0)
	if n := len(f.Pipes()); n != 2 {
		t.Fatalf("%d 个 tick 后应有 2 根管子, 实际 %d", PipeInterval, n)
	}
}

func TestPipeFieldScoresOncePerPipe(t *testing.T) {
	f := NewPipeField(7, DifficultyNormal)
	f.timer = math.MaxInt // 关掉自动生成, 只留下开局那根管子

	for i := 0; i < 1000; i++ {
		f.Update(PlayerX)
	}

	if got := f.Score(); got != 1 {
		t.Fatalf("同一根管子只应记 1 分, 实际 %d", got)
	}
	if n := len(f.Pipes()); n != 0 {
		t.Fatalf("移出屏幕的管子应当被回收, 还剩 %d 根", n)
	}
}

func TestPipeFieldGapChangeIsLimited(t *testing.T) {
	f := NewPipeField(9, DifficultyNormal)
	center := func(p *Pipe) float64 { return p.GapY + PipeGapHeight/2 }
	prev := center(f.Pipes()[0])

	for i := 0; i < 30; i++ {
		f.pipes = nil // 只关心新生成的这根管子
		f.spawn()
		cur := center(f.pipes[0])
		if math.Abs(cur-prev) > MaxGapDelta {
			t.Fatalf("相邻管子空隙中心变化 %v，超过上限 %v", math.Abs(cur-prev), MaxGapDelta)
		}
		prev = cur
	}
}

func TestHealthHurtAndInvulnerability(t *testing.T) {
	h := NewHealth(MaxHearts)
	if h.Hearts() != MaxHearts || h.Max() != MaxHearts || h.Dead() {
		t.Fatalf("满血状态不正确: hearts=%d max=%d", h.Hearts(), h.Max())
	}

	if !h.Hurt() {
		t.Fatal("满血时第一次碰撞应当扣血")
	}
	if got := h.Hearts(); got != MaxHearts-1 {
		t.Fatalf("受伤后剩余 %d 颗心, 期望 %d", got, MaxHearts-1)
	}
	if !h.Invulnerable() {
		t.Fatal("受伤后应当进入无敌时间")
	}

	// 无敌时间内无论被判定多少次碰撞，都只算一次伤害。
	for i := 0; i < InvulnTicks-1; i++ {
		h.Update()
		if h.Hurt() {
			t.Fatalf("无敌时间内不应再次扣血（第 %d tick）", i+1)
		}
	}
	if !h.Invulnerable() {
		t.Fatal("倒计时未走完时仍应处于无敌状态")
	}

	h.Update()
	if h.Invulnerable() {
		t.Fatal("经过 InvulnTicks 个 tick 后无敌时间应当结束")
	}
	if !h.Hurt() {
		t.Fatal("无敌时间结束后应当可以再次扣血")
	}
	if got := h.Hearts(); got != MaxHearts-2 {
		t.Fatalf("第二次受伤后剩余 %d 颗心, 期望 %d", got, MaxHearts-2)
	}
}

func TestHealthDiesWhenHeartsRunOut(t *testing.T) {
	h := NewHealth(2)

	for i := 0; i < 2; i++ {
		if !h.Hurt() {
			t.Fatalf("第 %d 次受伤应当生效", i+1)
		}
		for j := 0; j < InvulnTicks; j++ {
			h.Update()
		}
	}

	if !h.Dead() {
		t.Fatal("两颗心都扣完后应当判定为死亡")
	}
	if got := h.Hearts(); got != 0 {
		t.Fatalf("死亡时剩余 %d 颗心, 期望 0", got)
	}
	if h.Hurt() {
		t.Fatal("死亡后不应再扣血")
	}
}

// newTestPlayScene 创建一个没有管子干扰的游戏画面，只用来测试天花板/地板受伤逻辑。
func newTestPlayScene(hearts int) *playScene {
	s := &playScene{
		player: NewPlayer(60, 75),
		pipes:  NewPipeField(1, DifficultyNormal),
		health: NewHealth(hearts),
	}
	s.pipes.pipes = nil
	s.pipes.timer = math.MaxInt
	return s
}

// newTestGame 创建一个只带音效的游戏实例（音效播放器全是 nil，播放是空操作）。
func newTestGame() *Game {
	return &Game{sfx: &sound.Sound{}}
}

func TestPlaySceneCeilingHurtsAndClamps(t *testing.T) {
	s := newTestPlayScene(3)
	s.player.Y = 1
	s.player.VelY = -5

	if err := s.Update(newTestGame()); err != nil {
		t.Fatal(err)
	}

	if got := s.health.Hearts(); got != 2 {
		t.Fatalf("撞到天花板后剩余 %d 颗心, 期望 2", got)
	}
	if s.player.Y != 0 {
		t.Fatalf("角色应被贴回屏幕顶边, y = %v", s.player.Y)
	}
	if s.player.VelY != 0 {
		t.Fatalf("撞到天花板后竖直速度应清零, VelY = %v", s.player.VelY)
	}
}

func TestPlaySceneFloorDrainsHeartsThenGameOver(t *testing.T) {
	s := newTestPlayScene(3)
	g := newTestGame()

	for i := 0; i < 600 && s.health.Hearts() == 3; i++ {
		if err := s.Update(g); err != nil {
			t.Fatal(err)
		}
		if s.player.Rect().Bottom() > ScreenHeight {
			t.Fatalf("角色不应跑出屏幕底部, y = %v", s.player.Y)
		}
	}
	if got := s.health.Hearts(); got != 2 {
		t.Fatalf("掉到地板后剩余 %d 颗心, 期望 2", got)
	}

	// 一直不操作，角色会一直贴着地板，但无敌时间保证每 60 tick 才扣一次血。
	for i := 0; i < 600; i++ {
		if _, over := g.scene.(*gameOverScene); over {
			break
		}
		if err := s.Update(g); err != nil {
			t.Fatal(err)
		}
	}

	if !s.health.Dead() {
		t.Fatal("血量应当已经耗尽")
	}
	if _, ok := g.scene.(*gameOverScene); !ok {
		t.Fatalf("血量耗尽后应当切换到结束画面, 实际场景类型 %T", g.scene)
	}
}

func TestHealthShieldAbsorbsCollisions(t *testing.T) {
	h := NewHealth(2)
	h.AddShield(ItemHelmetCharges)
	if got := h.ShieldCharges(); got != ItemHelmetCharges {
		t.Fatalf("头盔层数 %d, 期望 %d", got, ItemHelmetCharges)
	}

	// 有护盾时碰撞只消耗护盾，不扣心血。
	for i := 0; i < ItemHelmetCharges; i++ {
		if !h.Hurt() {
			t.Fatalf("第 %d 次碰撞应当生效", i+1)
		}
		if got := h.Hearts(); got != 2 {
			t.Fatalf("有头盔时不应扣心, 剩余 %d 颗", got)
		}
		for j := 0; j < InvulnTicks; j++ {
			h.Update()
		}
	}
	if got := h.ShieldCharges(); got != 0 {
		t.Fatalf("护盾应当耗尽, 剩余 %d 层", got)
	}

	// 护盾用完之后的碰撞才扣心。
	if !h.Hurt() {
		t.Fatal("护盾耗尽后的碰撞应当扣心")
	}
	if got := h.Hearts(); got != 1 {
		t.Fatalf("护盾耗尽后应扣 1 颗心, 剩余 %d 颗", got)
	}
}

func TestHealthAddShieldCaps(t *testing.T) {
	h := NewHealth(MaxHearts)
	h.AddShield(2)
	h.AddShield(5)
	if got := h.ShieldCharges(); got != MaxShield {
		t.Fatalf("护盾层数应封顶在 %d, 实际 %d", MaxShield, got)
	}
}

func TestHealthHealCapsAtMax(t *testing.T) {
	h := NewHealth(MaxHearts)
	if got := h.Heal(3); got != 0 {
		t.Fatalf("满血时回血不应生效, 实际恢复 %d 颗", got)
	}

	h.Hurt()
	for i := 0; i < InvulnTicks; i++ {
		h.Update()
	}
	if got := h.Heal(5); got != 1 {
		t.Fatalf("应当只恢复 1 颗心, 实际 %d 颗", got)
	}
	if got := h.Hearts(); got != MaxHearts {
		t.Fatalf("回血后应回到上限 %d, 实际 %d", MaxHearts, got)
	}
}

func TestPlayerShrinkClamps(t *testing.T) {
	p := NewPlayer(60, 75)
	cx, cy := p.X+p.W/2, p.Y+p.H/2

	p.SetScale(ItemShrinkScale)

	wantW, wantH := 60*ItemShrinkScale, 75*ItemShrinkScale
	r := p.Rect()
	if math.Abs(r.W-wantW) > 1e-9 || math.Abs(r.H-wantH) > 1e-9 {
		t.Fatalf("缩小后碰撞盒 %vx%v, 期望 %vx%v", r.W, r.H, wantW, wantH)
	}
	// 缩放时中心应当保持不动。
	if math.Abs(r.X+r.W/2-cx) > 1e-9 || math.Abs(r.Y+r.H/2-cy) > 1e-9 {
		t.Fatalf("缩放后中心 (%v,%v) 偏离原中心 (%v,%v)", r.X+r.W/2, r.Y+r.H/2, cx, cy)
	}

	// 任何道具都不能把角色缩得比下限更小。
	p.SetScale(0.01)
	if p.Scale != PlayerMinScale {
		t.Fatalf("缩放比例应停在下限 %v, 实际 %v", PlayerMinScale, p.Scale)
	}
}

func TestPlayerShrinkExpires(t *testing.T) {
	p := NewPlayer(60, 75)
	p.Shrink(ItemShrinkTicks)
	if p.Scale != ItemShrinkScale {
		t.Fatalf("拾取缩小道具后比例 %v, 期望 %v", p.Scale, ItemShrinkScale)
	}

	// 持续时间还没到，仍然是缩小的。
	for i := 0; i < ItemShrinkTicks-1; i++ {
		p.Update()
	}
	if p.Scale != ItemShrinkScale {
		t.Fatal("持续时间还没到就不应该恢复")
	}

	// 时间到，恢复原尺寸与碰撞盒。
	p.Update()
	if p.Scale != 1 {
		t.Fatalf("缩小效果到期后应恢复原尺寸, 实际比例 %v", p.Scale)
	}
	if r := p.Rect(); math.Abs(r.W-60) > 1e-9 || math.Abs(r.H-75) > 1e-9 {
		t.Fatalf("恢复后碰撞盒 %vx%v, 期望 60x75", r.W, r.H)
	}
}

func TestPlayerShrinkRefreshes(t *testing.T) {
	p := NewPlayer(60, 75)
	p.Shrink(ItemShrinkTicks)
	for i := 0; i < ItemShrinkTicks-10; i++ {
		p.Update()
	}

	// 缩小时再吃一个，应该重新计时而不是立刻恢复。
	p.Shrink(ItemShrinkTicks)
	for i := 0; i < ItemShrinkTicks-1; i++ {
		p.Update()
	}
	if p.Scale != ItemShrinkScale {
		t.Fatal("重复拾取缩小道具应当刷新持续时间")
	}
}

func TestPipeFieldCarriesAndCollectsItems(t *testing.T) {
	f := NewPipeField(21, DifficultyNormal)
	f.spawnItem = func(speed float64) (ItemKind, bool) { return ItemHelmet, true }
	f.pipes = nil
	f.spawn()

	p := f.Pipes()[0]
	it := p.Item()
	if it == nil {
		t.Fatal("管子应当携带道具")
	}
	if it.Kind != ItemHelmet {
		t.Fatalf("道具种类 %v, 期望头盔", it.Kind)
	}

	// 道具放在空隙中心（水平居中在管子上）。
	if gapCenter := p.GapY + PipeGapHeight/2; math.Abs(it.Y+ItemSize/2-gapCenter) > 1e-9 {
		t.Fatalf("道具未垂直居中于空隙: Y=%v, 空隙中心=%v", it.Y+ItemSize/2, gapCenter)
	}
	if wantX := p.X + (PipeWidth-ItemSize)/2; math.Abs(it.X-wantX) > 1e-9 {
		t.Fatalf("道具未水平居中于管子: X=%v, 期望 %v", it.X, wantX)
	}

	// 道具跟随管子一起向左移动。
	xBefore := it.X
	f.Update(0)
	if it.X >= xBefore {
		t.Fatalf("道具应当随管子向左移动, %v -> %v", xBefore, it.X)
	}

	// 拾取后从管子上移除，不会重复拾取。
	got := f.CollectItems(it.Rect())
	if len(got) != 1 || got[0] != ItemHelmet {
		t.Fatalf("拾取结果 %v, 期望 [头盔]", got)
	}
	if p.Item() != nil {
		t.Fatal("拾取后道具应当被移除")
	}
	if again := f.CollectItems(it.Rect()); len(again) != 0 {
		t.Fatalf("道具不应被重复拾取, 实际 %v", again)
	}
}

// newItemPickupScene 创建一个角色正踩在道具上的场景，用来测试「拾取 → 生效」的整条链路。
func newItemPickupScene(kind ItemKind) *playScene {
	s := newTestPlayScene(MaxHearts)
	s.pipes.timer = math.MaxInt
	s.pipes.pipes = []*Pipe{{
		X:    ScreenWidth, // 管子本体远离角色，不会产生碰撞干扰
		GapY: 100,         // 空隙 100~250，正好包住角色
		item: &Item{X: s.player.X, Y: s.player.Y, Kind: kind},
	}}
	return s
}

func TestPlaySceneAppliesPickedItems(t *testing.T) {
	// 头盔：立刻获得护盾层数，道具从管子上消失。
	s := newItemPickupScene(ItemHelmet)
	if err := s.Update(newTestGame()); err != nil {
		t.Fatal(err)
	}
	if got := s.health.ShieldCharges(); got != ItemHelmetCharges {
		t.Fatalf("拾取头盔后护盾层数 %d, 期望 %d", got, ItemHelmetCharges)
	}
	if it := s.pipes.Pipes()[0].Item(); it != nil {
		t.Fatal("拾取后道具应当从管子上移除")
	}

	// 缩小：角色体型变小。
	s = newItemPickupScene(ItemShrink)
	if err := s.Update(newTestGame()); err != nil {
		t.Fatal(err)
	}
	if s.player.Scale != ItemShrinkScale {
		t.Fatalf("拾取缩小道具后缩放比例 %v, 期望 %v", s.player.Scale, ItemShrinkScale)
	}

	// 回血：补回失去的心。
	s = newItemPickupScene(ItemHeart)
	s.health.Hurt()
	if err := s.Update(newTestGame()); err != nil {
		t.Fatal(err)
	}
	if got := s.health.Hearts(); got != MaxHearts {
		t.Fatalf("拾取回血道具后应回满, 实际 %d 颗", got)
	}

	// 减速：卷动速度下降。
	s = newItemPickupScene(ItemSlow)
	before := s.pipes.Speed()
	if err := s.Update(newTestGame()); err != nil {
		t.Fatal(err)
	}
	if got := s.pipes.Speed(); got >= before {
		t.Fatalf("拾取减速道具后速度应下降: %v -> %v", before, got)
	}
}

func TestPipeFieldSlowDown(t *testing.T) {
	f := NewPipeField(2, DifficultyNormal)
	f.ticks = 3600 // 存活一分钟，速度已经明显高于开局
	base := f.Speed()

	f.SlowDown(ItemSlowAmount)
	if got := f.Speed(); math.Abs(got-(base-ItemSlowAmount)) > 1e-9 {
		t.Fatalf("减速后速度 %v, 期望 %v", got, base-ItemSlowAmount)
	}

	// 减速有下限，不会把速度压到停滞。
	f.SlowDown(100)
	if got := f.Speed(); got != ScrollSpeedFloor {
		t.Fatalf("减速下限应为 %v, 实际 %v", ScrollSpeedFloor, got)
	}
}

// newTestItemScene 创建一个只用于测试道具规则的最小场景（不需要图形环境）。
func newTestItemScene(seed uint64) *playScene {
	return &playScene{
		player: NewPlayer(60, 75),
		pipes:  NewPipeField(seed, DifficultyNormal),
		health: NewHealth(MaxHearts),
		rng:    rand.New(rand.NewPCG(seed, seed>>1)),
	}
}

func TestPickItemRespectsConditions(t *testing.T) {
	// 速度没到门槛时不刷减速道具；满血时不刷回血道具。
	s := newTestItemScene(1)
	for i := 0; i < 5000; i++ {
		kind, ok := s.pickItem(ItemSlowMinSpeed - 0.1)
		if !ok {
			continue
		}
		switch kind {
		case ItemSlow:
			t.Fatal("速度未达标时不应刷出减速道具")
		case ItemHeart:
			t.Fatal("满血时不应刷出回血道具")
		}
	}

	// 头盔层数已满时不再刷头盔；缩小道具则仍然可以出现（用来刷新持续时间）。
	s.player.Shrink(ItemShrinkTicks)
	s.health.AddShield(MaxShield)
	sawShrink := false
	for i := 0; i < 5000; i++ {
		kind, ok := s.pickItem(ScrollSpeed)
		if !ok {
			continue
		}
		if kind == ItemHelmet {
			t.Fatal("头盔层数已满时不应再刷出头盔")
		}
		if kind == ItemShrink {
			sawShrink = true
		}
	}
	if !sawShrink {
		t.Fatal("缩小状态下缩小道具仍应出现，用于刷新持续时间")
	}
}

func TestPickItemCanSpawnEveryKind(t *testing.T) {
	s := newTestItemScene(7)
	s.health.Hurt() // 掉一颗心，留出回血空间

	seen := map[ItemKind]bool{}
	for i := 0; i < 5000; i++ {
		if kind, ok := s.pickItem(ItemSlowMinSpeed + 0.5); ok {
			seen[kind] = true
		}
	}

	for _, kind := range []ItemKind{ItemHeart, ItemHelmet, ItemShrink, ItemSlow} {
		if !seen[kind] {
			t.Fatalf("道具 %v 在 5000 次生成中从未出现", kind)
		}
	}
}

func TestDifficultySpeedCurves(t *testing.T) {
	normal, hard := DifficultyNormal, DifficultyHard

	// 两种难度都从同样的速度起步。
	if got := normal.SpeedAt(0); got != ScrollSpeed {
		t.Fatalf("普通模式开局速度 %v, 期望 %v", got, ScrollSpeed)
	}
	if got := hard.SpeedAt(0); got != ScrollSpeed {
		t.Fatalf("困难模式开局速度 %v, 期望 %v", got, ScrollSpeed)
	}

	// 困难模式提速更快：同样的存活时间速度更高，但不会超过上限。
	for _, ticks := range []int{600, 1800, 3600} {
		if hard.SpeedAt(ticks) <= normal.SpeedAt(ticks) {
			t.Fatalf("%d tick 时困难模式速度 %v 应当高于普通模式 %v",
				ticks, hard.SpeedAt(ticks), normal.SpeedAt(ticks))
		}
	}

	// 普通上限不变，困难上限是普通的 1.5 倍，且再久也不会超过。
	if got := normal.SpeedAt(1_000_000); got != ScrollSpeedMax {
		t.Fatalf("普通模式上限 %v, 期望 %v", got, ScrollSpeedMax)
	}
	hardMax := ScrollSpeedMax * HardSpeedMaxScale
	if got := hard.SpeedAt(1_000_000); got != hardMax {
		t.Fatalf("困难模式上限 %v, 期望 %v", got, hardMax)
	}
	if hardMax > ScrollSpeedMax*1.5 {
		t.Fatalf("困难模式上限 %v 超过普通上限的 1.5 倍", hardMax)
	}

	// 两头对齐：普通模式到不了困难模式的上限。
	if normal.SpeedAt(1_000_000) >= hardMax {
		t.Fatal("普通模式的速度不应达到困难模式的上限")
	}
}

func TestDifficultyButtonsLayout(t *testing.T) {
	if numDifficulties < 2 {
		t.Fatalf("至少需要两个难度, 实际 %d", numDifficulties)
	}

	prevRight := 0.0
	for i := 0; i < numDifficulties; i++ {
		r := difficultyButtonRect(i)
		if r.X < prevRight {
			t.Fatalf("按钮 %d 与前一个按钮重叠", i)
		}
		if r.X < 0 || r.Right() > ScreenWidth {
			t.Fatalf("按钮 %d 超出画面: x=%v..%v", i, r.X, r.Right())
		}
		prevRight = r.Right()

		// 按钮中心应当命中自己。
		if d, ok := hitTestDifficulty(r.X+r.W/2, r.Y+r.H/2); !ok || d != Difficulty(i) {
			t.Fatalf("按钮 %d 的中心没有命中自己", i)
		}
	}

	if _, ok := hitTestDifficulty(0, 0); ok {
		t.Fatal("按钮之外的点击不应命中任何难度")
	}
}

func TestTitleDifficultySelection(t *testing.T) {
	s := newTitleScene()
	g := newTestGame()

	if s.selected != DifficultyNormal {
		t.Fatalf("默认应当选中普通模式, 实际 %v", s.selected)
	}

	// 向左循环到最后一个难度，再向右回到普通。
	g.input.menuDelta = -1
	if err := s.Update(g); err != nil {
		t.Fatal(err)
	}
	if s.selected != DifficultyHard {
		t.Fatalf("向左应当循环到困难模式, 实际 %v", s.selected)
	}
	g.input.menuDelta = 1
	if err := s.Update(g); err != nil {
		t.Fatal(err)
	}
	if s.selected != DifficultyNormal {
		t.Fatalf("向右应当回到普通模式, 实际 %v", s.selected)
	}

	// 点中「困难」按钮时应当用困难模式开始。
	r := difficultyButtonRect(int(DifficultyHard))
	g.input.hasPressPos = true
	g.input.pressX, g.input.pressY = r.X+r.W/2, r.Y+r.H/2
	if got := s.startDifficulty(g.input); got != DifficultyHard {
		t.Fatalf("点击困难按钮应返回困难模式, 实际 %v", got)
	}

	// 键盘开始时沿用当前选中的难度。
	g.input.hasPressPos = false
	if got := s.startDifficulty(g.input); got != s.selected {
		t.Fatalf("未点中按钮时应沿用当前选择, 实际 %v, 期望 %v", got, s.selected)
	}
}

func TestDisplayButtonsLayout(t *testing.T) {
	if numDisplayModes < 2 {
		t.Fatalf("至少需要窗口和全屏两种模式, 实际 %d", numDisplayModes)
	}

	prevRight := 0.0
	for i := 0; i < numDisplayModes; i++ {
		r := displayButtonRect(i)
		if r.X < prevRight {
			t.Fatalf("显示模式按钮 %d 与前一个按钮重叠", i)
		}
		if r.X < 0 || r.Right() > ScreenWidth {
			t.Fatalf("显示模式按钮 %d 超出画面: x=%v..%v", i, r.X, r.Right())
		}
		if r.Y < difficultyButtonRect(0).Bottom() {
			t.Fatalf("显示模式按钮 %d 与难度按钮重叠", i)
		}
		prevRight = r.Right()

		// 按钮中心应当命中自己。
		if m, ok := hitTestDisplay(r.X+r.W/2, r.Y+r.H/2); !ok || m != DisplayMode(i) {
			t.Fatalf("显示模式按钮 %d 的中心没有命中自己", i)
		}
	}

	if _, ok := hitTestDisplay(0, 0); ok {
		t.Fatal("按钮之外的点击不应命中任何显示模式")
	}
}

func TestTitleDisplaySelection(t *testing.T) {
	g := newTestGame()
	if g.Display() != DisplayWindowed {
		t.Fatalf("默认应当是窗口模式, 实际 %v", g.Display())
	}

	// 点中 FULLSCREEN 按钮时应当返回该模式。
	r := displayButtonRect(int(DisplayFullscreen))
	g.input.hasPressPos = true
	g.input.pressX, g.input.pressY = r.X+r.W/2, r.Y+r.H/2
	m, ok := displayChoice(g.input)
	if !ok || m != DisplayFullscreen {
		t.Fatalf("点击 FULLSCREEN 按钮应返回全屏模式, 实际 %v (ok=%v)", m, ok)
	}

	// 点击显示模式按钮只切换设置，不应开始游戏。
	s := newTitleScene()
	g.input.jumpPressed = true
	if err := s.Update(g); err != nil {
		t.Fatal(err)
	}
	if g.Display() != DisplayFullscreen {
		t.Fatalf("点击后应切换到全屏模式, 实际 %v", g.Display())
	}
	if g.scene != nil {
		t.Fatal("点击显示模式按钮不应开始游戏")
	}

	// 按钮之外的点击不是显示模式选择。
	g.input.pressX, g.input.pressY = 0, 0
	if _, ok := displayChoice(g.input); ok {
		t.Fatal("按钮之外的点击不应返回显示模式")
	}
}

func TestScrollSpeedRampsLogarithmically(t *testing.T) {
	if got := DifficultyNormal.SpeedAt(0); got != ScrollSpeed {
		t.Fatalf("开局速度 %v, 期望 %v", got, ScrollSpeed)
	}

	prev := DifficultyNormal.SpeedAt(0)
	for ticks := 100; ticks <= 20000; ticks += 100 {
		cur := DifficultyNormal.SpeedAt(ticks)
		if cur < prev {
			t.Fatalf("速度不应随时间下降: %d tick 时 %v < %v", ticks, cur, prev)
		}
		if cur > ScrollSpeedMax {
			t.Fatalf("%d tick 时速度 %v 超过上限 %v", ticks, cur, ScrollSpeedMax)
		}
		prev = cur
	}

	// 对数增长：同样长的存活区间，越靠后增量越小。
	early := DifficultyNormal.SpeedAt(1200) - DifficultyNormal.SpeedAt(0)
	late := DifficultyNormal.SpeedAt(2400) - DifficultyNormal.SpeedAt(1200)
	if late >= early {
		t.Fatalf("增速应逐渐放缓: 前 1200 tick 增加 %v, 后 1200 tick 增加 %v", early, late)
	}

	// 存活再久也不会超过上限。
	if got := DifficultyNormal.SpeedAt(1_000_000); got != ScrollSpeedMax {
		t.Fatalf("超长时间存活后速度应为上限 %v, 实际 %v", ScrollSpeedMax, got)
	}
}

func TestPipeFieldSpeedsUpOverTime(t *testing.T) {
	f := NewPipeField(11, DifficultyNormal)
	f.timer = math.MaxInt

	pipe := f.Pipes()[0]
	f.Update(0)
	early := ScreenWidth - pipe.X

	// 让整片管子区域已经「存活」很久，再生成一根新管子观察它的速度。
	f.pipes = nil
	f.ticks = 20000
	f.spawn()
	latePipe := f.Pipes()[0]
	f.Update(0)
	late := ScreenWidth - latePipe.X

	if late <= early {
		t.Fatalf("长时间存活后管子应移动更快: 开局 %v 像素/tick, 后期 %v 像素/tick", early, late)
	}
	if late > ScrollSpeedMax {
		t.Fatalf("后期速度 %v 超过上限 %v", late, ScrollSpeedMax)
	}
}

func TestPipeFieldCollides(t *testing.T) {
	f := NewPipeField(3, DifficultyNormal)
	f.timer = math.MaxInt
	p := f.Pipes()[0]

	tests := []struct {
		name string
		body Rect
		want bool
	}{
		{"撞到上半根管子", Rect{X: p.X + 1, Y: p.GapY - 10, W: 10, H: 10}, true},
		{"撞到下半根管子", Rect{X: p.X + 1, Y: p.GapY + PipeGapHeight + 5, W: 10, H: 10}, true},
		{"从空隙中穿过", Rect{X: p.X + 1, Y: p.GapY + 5, W: 10, H: 10}, false},
		{"还在管子左侧", Rect{X: p.X - 20, Y: p.GapY - 10, W: 10, H: 10}, false},
		{"在管子右侧", Rect{X: p.X + PipeWidth + 1, Y: p.GapY - 10, W: 10, H: 10}, false},
	}

	for _, tt := range tests {
		if got := f.Collides(tt.body); got != tt.want {
			t.Errorf("%s: Collides() = %v, 期望 %v", tt.name, got, tt.want)
		}
	}
}

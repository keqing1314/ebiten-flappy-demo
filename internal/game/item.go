package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/keqing1314/ebiten-flappy-demo/internal/assets"
)

// ItemKind 是道具的种类。
//
// 新增道具只需要三步：在这里加一个常量、把贴图放进 assets、在 itemSpecs 里加一条描述。
// 生成逻辑、绘制、拾取流程都不需要改。
type ItemKind int

const (
	ItemHeart  ItemKind = iota // 恢复一颗心
	ItemHelmet                 // 头盔：免疫若干次碰撞
	ItemShrink                 // 缩小角色
	ItemSlow                   // 降低卷动速度
)

// Item 是浮在管子空隙中心、等着被拾取的道具。
//
// 它挂在某根管子上，跟着管子一起向左移动；管子被回收时道具也一起消失，
// 因此不需要单独的生命周期管理。(X, Y) 是道具碰撞盒的左上角。
type Item struct {
	X, Y float64
	Kind ItemKind
}

// Rect 返回道具的碰撞盒。
func (it *Item) Rect() Rect {
	return Rect{X: it.X, Y: it.Y, W: ItemSize, H: ItemSize}
}

// Draw 把道具图标缩放后画在 (X, Y)。
func (it *Item) Draw(dst, img *ebiten.Image) {
	drawSpriteScaled(dst, img, it.X, it.Y, ItemSize, 1)
}

// itemSpec 描述一种道具：出现概率、出现条件、贴图与拾取效果。
type itemSpec struct {
	Kind ItemKind
	// Weight 是相对权重，只在「当前满足出现条件」的道具之间比较，越大越常见。
	Weight int
	// MinSpeed 是出现所需的最低卷动速度（像素/tick），0 表示不限制。
	MinSpeed float64
	// CanPick 是额外的出现条件（例如血满时不再刷回血道具），nil 表示没有额外条件。
	CanPick func(s *playScene) bool
	// Image 从资源集合里取出对应的贴图。
	Image func(a *assets.Assets) *ebiten.Image
	// Apply 把道具效果作用到当前这一局上。
	Apply func(s *playScene)
}

// itemSpecs 是所有道具的登记表。
var itemSpecs = []itemSpec{
	{
		Kind:   ItemHeart,
		Weight: 40,
		CanPick: func(s *playScene) bool {
			return s.health.Hearts() < s.health.Max()
		},
		Image: func(a *assets.Assets) *ebiten.Image { return a.ItemHeart },
		Apply: func(s *playScene) { s.health.Heal(1) },
	},
	{
		Kind:   ItemHelmet,
		Weight: 30,
		CanPick: func(s *playScene) bool {
			return s.health.ShieldCharges() < MaxShield
		},
		Image: func(a *assets.Assets) *ebiten.Image { return a.ItemHelmet },
		Apply: func(s *playScene) { s.health.AddShield(ItemHelmetCharges) },
	},
	{
		// 缩小时再捡到会在原来的基础上刷新持续时间，所以没有额外的出现条件。
		Kind:   ItemShrink,
		Weight: 20,
		Image:  func(a *assets.Assets) *ebiten.Image { return a.ItemShrink },
		Apply:  func(s *playScene) { s.player.Shrink(ItemShrinkTicks) },
	},
	{
		Kind:     ItemSlow,
		Weight:   10,
		MinSpeed: ItemSlowMinSpeed,
		Image:    func(a *assets.Assets) *ebiten.Image { return a.ItemSlow },
		Apply:    func(s *playScene) { s.pipes.SlowDown(ItemSlowAmount) },
	},
}

// itemSpecOf 返回某种道具的描述。
func itemSpecOf(kind ItemKind) *itemSpec {
	for i := range itemSpecs {
		if itemSpecs[i].Kind == kind {
			return &itemSpecs[i]
		}
	}
	return nil
}

// itemImage 返回某种道具的贴图。
func itemImage(a *assets.Assets, kind ItemKind) *ebiten.Image {
	if sp := itemSpecOf(kind); sp != nil {
		return sp.Image(a)
	}
	return nil
}

// canPick 报告这个道具当前能否出现。speed 是当前的卷动速度。
func (sp *itemSpec) canPick(s *playScene, speed float64) bool {
	if speed < sp.MinSpeed {
		return false
	}
	if sp.CanPick != nil && !sp.CanPick(s) {
		return false
	}
	return true
}

// pickItem 决定一根新管子要不要带道具、带哪一种。
//
// 先按 ItemSpawnChance 掷一次骰子，再在所有「当前可以出现」的道具中按权重抽一个。
// 如果一个都不满足条件（血满、已有头盔、已经缩小、速度还不够快……），这根管子就没有道具。
func (s *playScene) pickItem(speed float64) (ItemKind, bool) {
	if s.rng.Float64() >= ItemSpawnChance {
		return 0, false
	}

	total := 0
	for i := range itemSpecs {
		if itemSpecs[i].canPick(s, speed) {
			total += itemSpecs[i].Weight
		}
	}
	if total == 0 {
		return 0, false
	}

	roll := s.rng.IntN(total)
	for i := range itemSpecs {
		sp := &itemSpecs[i]
		if !sp.canPick(s, speed) {
			continue
		}
		if roll < sp.Weight {
			return sp.Kind, true
		}
		roll -= sp.Weight
	}
	return 0, false
}

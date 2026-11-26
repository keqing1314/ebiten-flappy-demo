package game

// Health 记录玩家的血量，以及受伤后的短暂无敌时间。
//
// 无敌时间（i-frames）是「可以挨好几次打」这类玩法必不可少的一环：
// 角色撞到管子后，管子还要继续向左移动几十个 tick 才能从角色身上移过去，
// 这期间每一 tick 都会被判定为碰撞。如果没有无敌时间，一次碰撞就会把 6 颗心全部扣光。
type Health struct {
	hearts int // 剩余血量
	max    int // 血量上限
	invuln int // 剩余的无敌时间（tick），0 表示可以被再次伤害
	shield int // 头盔提供的剩余免疫次数
}

// NewHealth 创建满血状态。
func NewHealth(max int) *Health {
	return &Health{hearts: max, max: max}
}

// Update 每个 tick 调用一次，让无敌时间倒计时。
func (h *Health) Update() {
	if h.invuln > 0 {
		h.invuln--
	}
}

// Hurt 处理一次碰撞：有护盾时消耗一层护盾，否则扣一颗心。
// 处于无敌时间中或已经死亡时返回 false，表示这次碰撞没有生效。
func (h *Health) Hurt() bool {
	if h.invuln > 0 || h.Dead() {
		return false
	}
	h.invuln = InvulnTicks
	if h.shield > 0 {
		h.shield--
		return true
	}
	h.hearts--
	return true
}

// AddShield 增加护盾层数（头盔道具），最多 MaxShield 层。
func (h *Health) AddShield(n int) {
	h.shield = min(h.shield+n, MaxShield)
}

// ShieldCharges 返回剩余护盾层数。
func (h *Health) ShieldCharges() int {
	return h.shield
}

// Heal 恢复 n 颗心（回血道具），不会超过上限。返回实际恢复的数量。
func (h *Health) Heal(n int) int {
	before := h.hearts
	h.hearts = min(h.hearts+n, h.max)
	return h.hearts - before
}

// Hearts 返回剩余血量。
func (h *Health) Hearts() int { return h.hearts }

// Max 返回血量上限。
func (h *Health) Max() int { return h.max }

// Dead 报告血量是否已经耗尽。
func (h *Health) Dead() bool { return h.hearts <= 0 }

// Invulnerable 报告当前是否处于受伤后的无敌时间。
func (h *Health) Invulnerable() bool { return h.invuln > 0 }

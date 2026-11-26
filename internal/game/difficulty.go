package game

import "math"

// Difficulty 是难度模式。不同模式只有卷动速度曲线不同：
//   - 普通模式：基础的对数曲线；
//   - 困难模式：从同样的速度起步，但提速快 3 倍，速度上限提高到普通上限的 1.5 倍。
//
// 上限是硬约束：无论玩多久，困难模式也不会快到玩家完全反应不过来。
type Difficulty int

const (
	DifficultyNormal Difficulty = iota
	DifficultyHard

	// numDifficulties 是难度总数，界面按它生成按钮、做循环选择。
	numDifficulties = int(DifficultyHard) + 1
)

// Name 返回界面上显示的名字（内置像素字体只有 ASCII 字形）。
func (d Difficulty) Name() string {
	if d == DifficultyHard {
		return "HARD"
	}
	return "NORMAL"
}

// SpeedAt 返回该难度下存活 ticks 个 tick 后的卷动速度（像素/tick）：
//
//	speed = min(上限, ScrollSpeed + ScrollSpeedGain*提速倍数*ln(1+ticks/ScrollSpeedRampTicks))
//
// 对数曲线保证了「前期提升明显、之后逐渐放缓」，上限则保证速度不会失控。
func (d Difficulty) SpeedAt(ticks int) float64 {
	speed := ScrollSpeed
	if ticks > 0 {
		speed += ScrollSpeedGain * d.gainScale() * math.Log1p(float64(ticks)/ScrollSpeedRampTicks)
	}
	if max := ScrollSpeedMax * d.maxScale(); speed > max {
		return max
	}
	return speed
}

// gainScale 返回该难度的提速倍数。
func (d Difficulty) gainScale() float64 {
	if d == DifficultyHard {
		return HardSpeedGainScale
	}
	return 1
}

// maxScale 返回该难度的速度上限相对普通上限的倍数。
func (d Difficulty) maxScale() float64 {
	if d == DifficultyHard {
		return HardSpeedMaxScale
	}
	return 1
}

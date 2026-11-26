package game

// Rect 是一个轴对齐矩形（AABB，Axis-Aligned Bounding Box）：
// X、Y 表示左上角，W、H 表示宽和高。
//
// 矩形碰撞是最常见的做法：计算量小、结果稳定，而且和像素美术的风格很搭。
// 游戏里所有的碰撞判定（地鼠 vs 管子）都建立在这个类型上。
type Rect struct {
	X, Y, W, H float64
}

// Right 返回右边缘的 x 坐标。
func (r Rect) Right() float64 { return r.X + r.W }

// Bottom 返回下边缘的 y 坐标。
func (r Rect) Bottom() float64 { return r.Y + r.H }

// Contains 报告点 (x, y) 是否落在矩形内（左闭右开，与 Intersects 的贴合语义一致）。
func (r Rect) Contains(x, y float64) bool {
	return x >= r.X && x < r.Right() && y >= r.Y && y < r.Bottom()
}

// Intersects 判断两个矩形是否相交。
//
// 这里用的是「分离轴」思想：只要在 x 方向或 y 方向上有一个方向不重叠，
// 两个矩形就不相交。注意比较用的是严格小于，
// 因此两个矩形「边缘刚刚贴合」不算碰撞，这样可以避免角色擦着管子边缘时被误判。
func (r Rect) Intersects(o Rect) bool {
	return r.X < o.Right() && o.X < r.Right() &&
		r.Y < o.Bottom() && o.Y < r.Bottom()
}

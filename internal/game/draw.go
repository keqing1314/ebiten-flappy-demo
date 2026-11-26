package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// 本文件是绘制用的小工具函数，被各个场景共用。

// drawBackground 把背景图铺在画布左上角。
// 背景图尺寸（640x360）与逻辑画布一致，所以直接 1:1 绘制即可。
func drawBackground(dst, bg *ebiten.Image) {
	dst.DrawImage(bg, nil)
}

// drawSprite 把贴图画在左上角坐标 (x, y) 处。
//
// DrawImageOptions 里的 GeoM 是一个 3x3 仿射变换矩阵，
// 平移、缩放、旋转都通过它完成；不做任何变换时传 nil 也可以。
func drawSprite(dst, img *ebiten.Image, x, y float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	dst.DrawImage(img, op)
}

// drawSpriteScaled 把贴图等比缩放到边长 size 后画在左上角 (x, y)。
// alpha 是不透明度（0~1）。
func drawSpriteScaled(dst, img *ebiten.Image, x, y, size, alpha float64) {
	scale := size / float64(img.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(img, op)
}

// drawText 在坐标 (x, y) 处绘制带阴影的文字。
// 阴影让文字在花哨的背景上依然清晰，是像素风界面常用的技巧。
func drawText(dst *ebiten.Image, s string, face *text.GoTextFace, x, y float64, clr color.Color) {
	// 先画一层向右下偏移的深色文字作为阴影。
	shadow := &text.DrawOptions{}
	shadow.GeoM.Translate(x+2, y+2)
	shadow.ColorScale.ScaleWithColor(color.RGBA{R: 0, G: 0, B: 0, A: 0xa0})
	text.Draw(dst, s, face, shadow)

	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, s, face, op)
}

// drawTextCenteredIn 在矩形 r 内水平、垂直居中地绘制文字。
func drawTextCenteredIn(dst *ebiten.Image, s string, face *text.GoTextFace, r Rect, clr color.Color) {
	w, h := text.Measure(s, face, face.Size)
	drawText(dst, s, face, r.X+(r.W-w)/2, r.Y+(r.H-h)/2, clr)
}

// drawTextCentered 水平居中地绘制文字，y 是文字顶部的坐标。
func drawTextCentered(dst *ebiten.Image, s string, face *text.GoTextFace, y float64, clr color.Color) {
	w, _ := text.Measure(s, face, face.Size)
	drawText(dst, s, face, (ScreenWidth-w)/2, y, clr)
}

// fillOverlay 在整个画面上盖一层半透明黑，用来「压暗」背景。
func fillOverlay(dst *ebiten.Image, alpha uint8) {
	vector.FillRect(dst, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{A: alpha}, false)
}

// optionButtonBorder 是标题画面选项按钮（难度、显示模式）共用的描边宽度。
const optionButtonBorder = 2.0

// drawOptionButton 绘制一个选项按钮：选中项用白底深色字（反色高亮），
// 未选中项压暗，保证两者都清晰可读。
func drawOptionButton(dst *ebiten.Image, r Rect, label string, face *text.GoTextFace, selected bool) {
	var fill, border, textColor color.Color = color.RGBA{A: 0x50},
		color.RGBA{R: 0x90, G: 0x90, B: 0x90, A: 0xff},
		color.RGBA{R: 0xa0, G: 0xa0, B: 0xa0, A: 0xff}
	if selected {
		fill = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		border = color.White
		textColor = color.RGBA{R: 0x20, G: 0x40, B: 0x60, A: 0xff}
	}

	vector.FillRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), fill, false)
	vector.StrokeRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), optionButtonBorder, border, false)
	drawTextCenteredIn(dst, label, face, r, textColor)
}

// strokeRect 用一个像素的线框描出矩形，仅在调试模式下用来显示碰撞盒。
func strokeRect(dst *ebiten.Image, r Rect, clr color.Color) {
	vector.StrokeRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), 1, clr, false)
}

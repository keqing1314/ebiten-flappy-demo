// Package assets 负责加载游戏需要的资源：贴图与字体。
//
// 资源文件通过 go:embed 编译进二进制文件，因此编译出来的单个 exe 就能独立运行，
// 不需要再把 png 文件拷来拷去（这也是 Ebitengine 项目常见的发布方式）。
package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// imageFS 把 images 目录下的所有 png 打进程序里。
// 注意：embed 只能包含当前目录及其子目录，所以资源文件必须放在本包内部。
//
//go:embed images/*.png
var imageFS embed.FS

// Assets 是加载完成的资源集合。
//
// 贴图在加载时就已经解码并上传到 GPU（*ebiten.Image），
// 之后每一帧只需要调用 DrawImage，开销很小。
type Assets struct {
	Sky    *ebiten.Image // 背景，640x360，正好铺满逻辑画布
	Wall   *ebiten.Image // 土管，21x361
	Gopher *ebiten.Image // 主角地鼠，60x75
	Heart  *ebiten.Image // 血条图标，1024x1024，绘制时按 HUD 尺寸缩放

	// 道具图标（都是 1024x1024，绘制时按 ItemSize 缩放）。
	ItemHeart  *ebiten.Image // 恢复一颗心
	ItemHelmet *ebiten.Image // 头盔：免疫若干次碰撞
	ItemShrink *ebiten.Image // 缩小角色
	ItemSlow   *ebiten.Image // 降低卷动速度

	gopherSource image.Image            // 未上传 GPU 的原始图像，用作窗口图标
	fontSource   *text.GoTextFaceSource // 字体源，可反复创建不同字号的样式
}

// Load 读取全部资源。任何一个资源缺失都会返回错误，方便尽早发现问题。
func Load() (*Assets, error) {
	dir, err := fs.Sub(imageFS, "images")
	if err != nil {
		return nil, fmt.Errorf("定位内嵌图片目录失败: %w", err)
	}

	a := &Assets{}

	// NewImageFromFileSystem 会解码 png 并生成可供 GPU 使用的 *ebiten.Image。
	// 第二个返回值是原始 image.Image，这里只有窗口图标用得上。
	if a.Sky, _, err = ebitenutil.NewImageFromFileSystem(dir, "sky.png"); err != nil {
		return nil, fmt.Errorf("加载 sky.png 失败: %w", err)
	}
	if a.Wall, _, err = ebitenutil.NewImageFromFileSystem(dir, "wall.png"); err != nil {
		return nil, fmt.Errorf("加载 wall.png 失败: %w", err)
	}
	if a.Gopher, a.gopherSource, err = ebitenutil.NewImageFromFileSystem(dir, "gopher.png"); err != nil {
		return nil, fmt.Errorf("加载 gopher.png 失败: %w", err)
	}
	if a.Heart, _, err = ebitenutil.NewImageFromFileSystem(dir, "heart.png"); err != nil {
		return nil, fmt.Errorf("加载 heart.png 失败: %w", err)
	}
	if a.ItemHeart, _, err = ebitenutil.NewImageFromFileSystem(dir, "item_heart.png"); err != nil {
		return nil, fmt.Errorf("加载 item_heart.png 失败: %w", err)
	}
	if a.ItemHelmet, _, err = ebitenutil.NewImageFromFileSystem(dir, "item_helmet.png"); err != nil {
		return nil, fmt.Errorf("加载 item_helmet.png 失败: %w", err)
	}
	if a.ItemShrink, _, err = ebitenutil.NewImageFromFileSystem(dir, "item_shrink.png"); err != nil {
		return nil, fmt.Errorf("加载 item_shrink.png 失败: %w", err)
	}
	if a.ItemSlow, _, err = ebitenutil.NewImageFromFileSystem(dir, "item_slow.png"); err != nil {
		return nil, fmt.Errorf("加载 item_slow.png 失败: %w", err)
	}

	// 使用 Ebitengine 官方示例自带的像素字体 PressStart2P。
	// 它只包含 ASCII 字形，所以界面上的文字使用英文（中文需要另外引入中文字体）。
	if a.fontSource, err = text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf)); err != nil {
		return nil, fmt.Errorf("解析字体失败: %w", err)
	}

	return a, nil
}

// NewFace 用内嵌字体创建一个指定字号（单位：像素）的文字样式。
// GoTextFace 本身很轻量，同一个字体源可以创建任意多个不同字号的样式。
func (a *Assets) NewFace(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: a.fontSource, Size: size}
}

// Icon 返回用于设置窗口图标的图像。
func (a *Assets) Icon() image.Image {
	return a.gopherSource
}

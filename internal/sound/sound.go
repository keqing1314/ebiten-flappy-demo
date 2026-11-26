// Package sound 负责游戏的全部音效。
//
// 拍翅膀、得分、失败三个短音效用代码现场合成：Ebitengine 的 audio 包播放的是
// 「原始 PCM 数据」（16 位有符号整数、双声道、左右声道交错、小端序），
// 只要把一串 []byte 交给它就能发声，于是这里用方波算出几个 8-bit 风格的效果音。
//
// 受伤音效则是随包内嵌的 mp3 文件（hurt.mp3），用 go:embed 打进二进制，
// 运行时由 audio/mp3 解码成 PCM 流再交给播放器——这也是加载音频文件的标准做法。
package sound

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

// hurtMP3 是受伤音效的原始字节。go:embed 只能嵌入当前目录及其子目录的文件。
//
//go:embed hurt.mp3
var hurtMP3 []byte

// SampleRate 是采样率：每秒采样多少个点。44100 是 CD 音质，足够用了。
// 合成 PCM 时的采样率必须和 audio.Context 的采样率一致，否则声音会变调。
const SampleRate = 44100

// Sound 持有所有音效播放器。
//
// 如果运行环境没有可用的音频设备（例如某些远程桌面、无音卡的服务器），
// Load 会返回一个所有字段都为 nil 的 Sound，此时所有播放方法都是空操作，游戏照常运行。
type Sound struct {
	flap  *audio.Player // 拍翅膀
	score *audio.Player // 得分
	hit   *audio.Player // 失败（血量耗尽）
	hurt  *audio.Player // 受伤（掉血）
}

// Load 合成音效、解码内嵌的 mp3，并创建全部播放器。
func Load() *Sound {
	ctx, err := newContext()
	if err != nil {
		fmt.Printf("音频设备不可用，已静音: %v\n", err)
		return &Sound{}
	}

	// NewPlayerFromBytes 会把整段 PCM 数据包成一个可重复播放的播放器。
	// 每个音效互不干扰，所以各自持有一个播放器。
	return &Sound{
		flap:  ctx.NewPlayerFromBytes(render(sweep(520, 900), 0.08, 0.35)),
		score: ctx.NewPlayerFromBytes(render(twoTone(880, 1320), 0.16, 0.35)),
		hit:   ctx.NewPlayerFromBytes(render(sweep(300, 90), 0.30, 0.50)),
		hurt:  newMP3Player(ctx, hurtMP3),
	}
}

// newMP3Player 把内嵌的 mp3 解码成可重复播放的播放器。
//
// mp3 是压缩格式，解码时要把采样率转换成与音频上下文一致，所以必须等 newContext 成功。
// 任何一步失败都返回 nil，调用方会静音处理，游戏照常运行。
func newMP3Player(ctx *audio.Context, data []byte) *audio.Player {
	stream, err := mp3.DecodeWithSampleRate(ctx.SampleRate(), bytes.NewReader(data))
	if err != nil {
		fmt.Printf("解码受伤音效失败，已跳过: %v\n", err)
		return nil
	}
	p, err := ctx.NewPlayer(stream)
	if err != nil {
		fmt.Printf("创建受伤音效播放器失败，已跳过: %v\n", err)
		return nil
	}
	return p
}

// Flap 播放「拍翅膀」音效。
func (s *Sound) Flap() { play(s.flap) }

// Score 播放「得分」音效。
func (s *Sound) Score() { play(s.score) }

// Hurt 播放「受伤」音效（掉一颗心时触发）。
func (s *Sound) Hurt() { play(s.hurt) }

// Hit 播放「失败」音效（血量耗尽时触发）。
func (s *Sound) Hit() { play(s.hit) }

// play 从头播放一个音效。
//
// 音效很短，同一个播放器播放结束后可以 Rewind 回到起点重复使用，
// 这样就不必每次播放都新建播放器（新建的对象如果被 GC 回收，声音会被中断）。
func play(p *audio.Player) {
	if p == nil {
		return
	}
	if err := p.Rewind(); err != nil {
		return
	}
	p.Play()
}

// newContext 创建音频上下文。
//
// audio.NewContext 是整个进程里只能调用一次的全局操作，并且在驱动初始化失败时会 panic，
// 所以这里用 recover 把它转换成一个普通错误，让游戏在无声环境下依然可以运行。
func newContext() (ctx *audio.Context, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return audio.NewContext(SampleRate), nil
}

// waveform 描述波形：输入 0..1 的播放进度，输出当前时刻的频率（Hz）。
// 把频率写成进度相关的函数，就能很方便地做滑音。
type waveform func(progress float64) float64

// sweep 返回一个音高从 from 线性滑到 to 的频率函数。
func sweep(from, to float64) waveform {
	return func(progress float64) float64 {
		return from + (to-from)*progress
	}
}

// twoTone 返回一个「前半段低、后半段高」的频率函数，常用于得分提示音。
func twoTone(low, high float64) waveform {
	return func(progress float64) float64 {
		if progress < 0.5 {
			return low
		}
		return high
	}
}

// render 按波形合成一段方波音效，返回可直接交给 audio 包播放的 PCM 数据。
func render(w waveform, seconds, volume float64) []byte {
	const bytesPerSample = 4 // 双声道 × 2 字节

	n := int(SampleRate * seconds)
	buf := make([]byte, n*bytesPerSample)

	var phase float64 // 累加相位，单位是「周期」
	for i := 0; i < n; i++ {
		progress := float64(i) / float64(n)

		// 累加相位而不是直接算 sin(2πft)：这样即使频率在变化，波形也是连续的。
		phase += w(progress) / SampleRate

		// 方波（取正弦的符号）听起来比正弦波更有复古游戏的味道。
		sample := math.Copysign(volume, math.Sin(2*math.Pi*phase))

		// 包络：开头快速淡入、结尾线性淡出，避免出现「咔哒」的爆音。
		envelope := math.Min(1, progress*40) * (1 - progress)

		v := int16(sample * envelope * math.MaxInt16)

		// 左右声道写入相同的采样值。
		binary.LittleEndian.PutUint16(buf[i*bytesPerSample:], uint16(v))
		binary.LittleEndian.PutUint16(buf[i*bytesPerSample+2:], uint16(v))
	}

	return buf
}

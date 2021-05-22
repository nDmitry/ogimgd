package preview

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/fogleman/gg"
)

const (
	margin         = 20.0
	padding        = 48.0
	border         = 8
	maxTitleLength = 90
)

// Options defines a set of options required to draw a p.ctx.
type Options struct {
	// Canvas width
	CanvasW int
	// Canvas height
	CanvasH int
	// Avatar diameter
	AvaD   int
	Title  string
	Author string
	// Logo left part text (optional)
	LabelL string
	// Logo right part text (required)
	LabelR string
	// An URL to a background image, either of canvas size or it will be thumbnailed and smart-cropped
	BgURL string
	// An URL to an author avatar pic
	AvaURL string
	// An URL to an icon image
	IconURL string
	// Icon width
	IconW int
	// Icon height
	IconH int
}

type Preview struct {
	Opts Options
	ctx  *gg.Context
}

func New(opts Options) *Preview {
	return &Preview{
		Opts: opts,
		ctx:  gg.NewContext(opts.CanvasW, opts.CanvasH),
	}
}

// Draw draws a preview using the options provided.
func (p *Preview) Draw() (image.Image, error) {
	if err := p.drawBackground(); err != nil {
		return nil, err
	}

	if err := p.drawForeground(); err != nil {
		return nil, err
	}

	if err := p.drawAvatar(); err != nil {
		return nil, err
	}

	if err := p.drawAuthor(); err != nil {
		return nil, err
	}

	if err := p.drawTitle(); err != nil {
		return nil, err
	}

	if err := p.drawLabel(); err != nil {
		return nil, err
	}

	return p.ctx.Image(), nil
}

func (p *Preview) drawBackground() error {
	bgBuf, err := os.ReadFile(p.Opts.BgURL)

	if err != nil {
		return fmt.Errorf("could not get the background %s: %w", p.Opts.BgURL, err)
	}

	bgBuf, err = resize(bgBuf, p.Opts.CanvasW, p.Opts.CanvasH)

	if err != nil {
		return fmt.Errorf("could not resize the background: %w", err)
	}

	bgImg, _, err := image.Decode(bytes.NewReader(bgBuf))

	if err != nil {
		return fmt.Errorf("could not decode the background: %w", err)
	}

	p.ctx.DrawImage(bgImg, 0, 0)

	return nil
}

func (p *Preview) drawForeground() error {
	p.ctx.SetColor(color.RGBA{0, 0, 0, 155})
	p.ctx.DrawRectangle(margin, margin, float64(p.Opts.CanvasW)-(margin*2), float64(p.Opts.CanvasH)-(margin*2))
	p.ctx.Fill()

	return nil
}

func (p *Preview) drawAvatar() error {
	// draw the avatar border circle
	avaX := padding + float64(p.Opts.AvaD+border)/2
	avaY := padding + float64(p.Opts.AvaD+border)/2

	p.ctx.DrawCircle(avaX, avaY, float64((p.Opts.AvaD+8)/2))
	p.ctx.SetHexColor("#FFFFFF")
	p.ctx.Fill()

	// draw the avatar itself (cropped to a circle)
	avaBuf, err := os.ReadFile(p.Opts.AvaURL)

	if err != nil {
		return fmt.Errorf("could not get the avatar %s: %w", p.Opts.AvaURL, err)
	}

	avaBuf, err = resize(avaBuf, p.Opts.AvaD, p.Opts.AvaD)

	if err != nil {
		return fmt.Errorf("could not resize the avatar: %w", err)
	}

	avaImg, _, err := image.Decode(bytes.NewReader(avaBuf))

	if err != nil {
		return fmt.Errorf("could not decode the avatar: %w", err)
	}

	avaImg = circle(avaImg)

	p.ctx.DrawImageAnchored(avaImg, int(avaX), int(avaY), 0.5, 0.5)

	return nil
}

func (p *Preview) drawAuthor() error {
	p.ctx.LoadFontFace(filepath.Join("fonts", "Ubuntu-Medium.ttf"), 36)
	p.ctx.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 204})

	authorX := padding + float64(p.Opts.AvaD) + padding/2
	authorY := padding + float64(p.Opts.AvaD)/2

	p.ctx.DrawStringAnchored(p.Opts.Author, authorX, authorY, 0, 0.5)

	return nil
}

func (p *Preview) drawTitle() error {
	p.ctx.LoadFontFace(filepath.Join("fonts", "Ubuntu-Medium.ttf"), 76)
	p.ctx.SetColor(color.White)

	titleX := padding
	titleY := padding*2 + float64(p.Opts.AvaD)
	maxWidth := float64(p.Opts.CanvasW) - padding - margin*2
	title := p.Opts.Title

	if utf8.RuneCountInString(title) > maxTitleLength {
		title = string([]rune(title)[0:maxTitleLength]) + "…"
	}

	p.ctx.DrawStringWrapped(title, titleX, titleY, 0, 0, maxWidth, 1.6, gg.AlignLeft)

	return nil
}

func (p *Preview) drawLabel() error {
	// draw the required right part of the label
	p.ctx.LoadFontFace(filepath.Join("fonts", "Ubuntu-Bold.ttf"), 36)
	p.ctx.SetColor(color.White)

	labelRightWidth, labelRightHeight := p.ctx.MeasureString(p.Opts.LabelR)
	labelX := float64(p.Opts.CanvasW) - labelRightWidth - padding
	labelY := float64(p.Opts.CanvasH) - padding

	p.ctx.DrawString(p.Opts.LabelR, labelX, labelY)

	// draw the icon
	iconBuf, err := os.ReadFile(p.Opts.IconURL)

	if err != nil {
		return fmt.Errorf("could not get the icon: %w", err)
	}

	iconBuf, err = resize(iconBuf, p.Opts.IconW, p.Opts.IconH)

	if err != nil {
		return fmt.Errorf("could not resize the icon: %w", err)
	}

	iconImg, _, err := image.Decode(bytes.NewReader(iconBuf))

	if err != nil {
		return fmt.Errorf("could not decode the icon: %w", err)
	}

	iconX := int(labelX) - p.Opts.IconW - int(margin/2)
	iconY := p.Opts.CanvasH - int(padding) - int(labelRightHeight/2)

	p.ctx.DrawImageAnchored(iconImg, iconX, iconY, 0, 0.5)

	if len(p.Opts.LabelL) > 0 {
		labelLeftWidth, _ := p.ctx.MeasureString(p.Opts.LabelL)

		labelX = float64(p.Opts.CanvasW) - labelLeftWidth - labelRightWidth - float64(p.Opts.IconW) - margin - padding
		labelY = float64(p.Opts.CanvasH) - padding

		p.ctx.DrawString(p.Opts.LabelL, labelX, labelY)
	}

	return nil
}

// resize resizes an image to the specified width and height if it differs from them.
// In case the aspect ratio of the source image differs from w/h parameters, it crops it to the area of interest.
func resize(buf []byte, w, h int) ([]byte, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(buf))

	if err != nil {
		return nil, err
	}

	if config.Width == w || config.Height == h {
		return buf, nil
	}

	log.Printf("Resizing an image to %dx%d px", w, h)

	vipsImg, err := vips.NewImageFromBuffer(buf)

	if err != nil {
		return nil, err
	}

	defer vipsImg.Close()

	if err = vipsImg.Thumbnail(w, h, vips.InterestingAttention); err != nil {
		return nil, err
	}

	buf, _, err = vipsImg.Export(vips.NewDefaultExportParams())

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// circle crops circle out of a rectangle source image.
func circle(src image.Image) image.Image {
	log.Printf("Circling an image")

	r := int(math.Min(
		float64(src.Bounds().Dx()),
		float64(src.Bounds().Dy()),
	) / 2)

	p := image.Point{
		X: src.Bounds().Dx() / 2,
		Y: src.Bounds().Dy() / 2,
	}

	mask := gg.NewContextForRGBA(image.NewRGBA(src.Bounds()))

	mask.DrawCircle(float64(p.X), float64(p.Y), float64(r))
	mask.Clip()
	mask.DrawImage(src, 0, 0)

	return mask.Image()
}

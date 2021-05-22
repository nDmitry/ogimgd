package preview

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"path/filepath"
	"unicode/utf8"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/fogleman/gg"
	"github.com/nDmitry/ogimgd/internal/remote"
)

const (
	margin         = 20.0
	padding        = 48.0
	border         = 8
	maxTitleLength = 90
)

type getter interface {
	GetAll(context.Context, []string) ([][]byte, error)
}

// Options defines a set of options required to draw a p.ctx.
type Options struct {
	// Canvas width
	CanvasW int
	// Canvas height
	CanvasH int
	// Opacity value for the black foreground under the title
	Opacity float64
	// Avatar diameter
	AvaD  int
	Title string
	// Title font size
	TitleSize float64
	Author    string
	// Author font size
	AuthorSize float64
	// Logo left part text (optional)
	LabelL string
	// Logo right part text (optional)
	LabelR string
	// Label font size
	LabelSize float64
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
	// Resulting JPEG quality
	Quality int
}

// Preview can draw a preview using the provided Options.
type Preview struct {
	opts   *Options
	ctx    *gg.Context
	remote getter
}

// New returns an initialized Preview.
func New() *Preview {
	return &Preview{
		opts:   nil,
		ctx:    nil,
		remote: remote.New(),
	}
}

// Draw draws a preview using the provided Options.
func (p *Preview) Draw(ctx context.Context, opts Options) (image.Image, error) {
	p.opts = &opts
	p.ctx = gg.NewContext(opts.CanvasW, opts.CanvasH)
	urlsOrPaths := []string{opts.BgURL, opts.AvaURL, opts.IconURL}

	imgBufs, err := p.remote.GetAll(ctx, urlsOrPaths)

	if err != nil {
		return nil, fmt.Errorf("could not get an image: %w", err)
	}

	if err := p.drawBackground(imgBufs[0]); err != nil {
		return nil, err
	}

	if err := p.drawForeground(); err != nil {
		return nil, err
	}

	if err := p.drawAvatar(imgBufs[1]); err != nil {
		return nil, err
	}

	if err := p.drawAuthor(); err != nil {
		return nil, err
	}

	if err := p.drawTitle(); err != nil {
		return nil, err
	}

	if err := p.drawLabel(imgBufs[2]); err != nil {
		return nil, err
	}

	return p.ctx.Image(), nil
}

func (p *Preview) drawBackground(bgBuf []byte) error {
	bgBuf, err := resize(bgBuf, p.opts.CanvasW, p.opts.CanvasH)

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
	p.ctx.SetColor(color.RGBA{0, 0, 0, uint8(255.0 * p.opts.Opacity)})
	p.ctx.DrawRectangle(margin, margin, float64(p.opts.CanvasW)-(margin*2), float64(p.opts.CanvasH)-(margin*2))
	p.ctx.Fill()

	return nil
}

func (p *Preview) drawAvatar(avaBuf []byte) error {
	// draw the avatar border circle
	avaX := padding + float64(p.opts.AvaD+border)/2
	avaY := padding + float64(p.opts.AvaD+border)/2

	p.ctx.DrawCircle(avaX, avaY, float64((p.opts.AvaD+8)/2))
	p.ctx.SetHexColor("#FFFFFF")
	p.ctx.Fill()

	// draw the avatar itself (cropped to a circle)
	avaBuf, err := resize(avaBuf, p.opts.AvaD, p.opts.AvaD)

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
	p.ctx.LoadFontFace(filepath.Join("fonts", "Ubuntu-Medium.ttf"), p.opts.AuthorSize)
	p.ctx.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 204})

	authorX := padding + float64(p.opts.AvaD) + padding/2
	authorY := padding + float64(p.opts.AvaD)/2

	p.ctx.DrawStringAnchored(p.opts.Author, authorX, authorY, 0, 0.5)

	return nil
}

func (p *Preview) drawTitle() error {
	p.ctx.LoadFontFace(filepath.Join("fonts", "Ubuntu-Medium.ttf"), p.opts.TitleSize)
	p.ctx.SetColor(color.White)

	titleX := padding
	titleY := padding*2 + float64(p.opts.AvaD)
	maxWidth := float64(p.opts.CanvasW) - padding - margin*2
	title := p.opts.Title

	if utf8.RuneCountInString(title) > maxTitleLength {
		title = string([]rune(title)[0:maxTitleLength]) + "…"
	}

	p.ctx.DrawStringWrapped(title, titleX, titleY, 0, 0, maxWidth, 1.6, gg.AlignLeft)

	return nil
}

func (p *Preview) drawLabel(iconBuf []byte) error {
	// draw the required right part of the label
	labelX := float64(p.opts.CanvasW) - padding
	labelY := float64(p.opts.CanvasH) - padding
	labelRightWidth := 0.0
	labelRightHeight := padding

	if len(p.opts.LabelL) > 0 {
		p.ctx.LoadFontFace(filepath.Join("fonts", "Ubuntu-Bold.ttf"), p.opts.LabelSize)
		p.ctx.SetColor(color.White)

		labelRightWidth, labelRightHeight = p.ctx.MeasureString(p.opts.LabelR)
		labelX -= labelRightWidth

		p.ctx.DrawString(p.opts.LabelR, labelX, labelY)
	}

	// draw the icon
	iconBuf, err := resize(iconBuf, p.opts.IconW, p.opts.IconH)

	if err != nil {
		return fmt.Errorf("could not resize the icon: %w", err)
	}

	iconImg, _, err := image.Decode(bytes.NewReader(iconBuf))

	if err != nil {
		return fmt.Errorf("could not decode the icon: %w", err)
	}

	iconX := int(labelX) - p.opts.IconW - int(margin/2)
	iconY := p.opts.CanvasH - int(padding) - int(labelRightHeight/2)

	p.ctx.DrawImageAnchored(iconImg, iconX, iconY, 0, 0.5)

	if len(p.opts.LabelL) > 0 {
		labelLeftWidth, _ := p.ctx.MeasureString(p.opts.LabelL)

		labelX = float64(p.opts.CanvasW) - labelLeftWidth - labelRightWidth - float64(p.opts.IconW) - margin - padding
		labelY = float64(p.opts.CanvasH) - padding

		p.ctx.DrawString(p.opts.LabelL, labelX, labelY)
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

	if config.Width == w && config.Height == h {
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

package preview

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"regexp"
	"unicode/utf8"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/fogleman/gg"
	"github.com/nDmitry/ogimgd/internal/remote"
)

const (
	margin            = 20.0
	padding           = 48.0
	border            = 8
	maxTitleLength    = 90
	defaultBgColor    = "#FFFFFF"
	avatarBorderColor = "#FFFFFF"
	logoKey           = "logo"
	avaKey            = "avatar"
	bgKey             = "bg"
)

var hexRe = regexp.MustCompile("^#(?:[0-9a-fA-F]{3}){1,2}$")

type getter interface {
	GetAll(context.Context, map[string]string) (map[string][]byte, error)
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
	// Either an URL to a remote background image, or filename of the local image, or a HEX-color
	// An image will be thumbnailed and smart-cropped if it's not of the canvas size
	Bg string
	// An URL to an author avatar pic
	AvaURL string
	// An URL to a logo image
	LogoURL string
	// Logo height
	LogoH int
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
	bgColor := defaultBgColor
	isBgHEX := hexRe.Match([]byte(p.opts.Bg))
	urlsOrPaths := map[string]string{logoKey: p.opts.LogoURL}

	if opts.AvaURL != "" {
		urlsOrPaths[avaKey] = p.opts.AvaURL
	}

	if isBgHEX {
		bgColor = p.opts.Bg
	} else if p.opts.Bg != "" {
		urlsOrPaths[bgKey] = p.opts.Bg
	}

	imgBufs, err := p.remote.GetAll(ctx, urlsOrPaths)

	if err != nil {
		return nil, fmt.Errorf("could not get an image: %w", err)
	}

	if isBgHEX || p.opts.Bg == "" {
		if err := p.drawBackground(nil, bgColor); err != nil {
			return nil, err
		}
	} else {
		if err := p.drawBackground(imgBufs[bgKey], bgColor); err != nil {
			return nil, err
		}
	}

	if err := p.drawForeground(); err != nil {
		return nil, err
	}

	if _, exists := imgBufs[avaKey]; exists {
		if err := p.drawAvatar(imgBufs[avaKey]); err != nil {
			return nil, err
		}
	}

	if err := p.drawAuthor(); err != nil {
		return nil, err
	}

	if err := p.drawTitle(); err != nil {
		return nil, err
	}

	if err := p.drawLogo(imgBufs[logoKey]); err != nil {
		return nil, err
	}

	return p.ctx.Image(), nil
}

func (p *Preview) drawBackground(bgBuf []byte, bgColor string) error {
	if bgBuf == nil {
		p.ctx.SetHexColor(bgColor)
		p.ctx.DrawRectangle(0, 0, float64(p.opts.CanvasW), float64(p.opts.CanvasH))
		p.ctx.Fill()

		return nil
	}

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
	p.ctx.SetHexColor(avatarBorderColor)
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
	if p.opts.Author == "" {
		return nil
	}

	font, err := loadFont(p.opts.AuthorSize)

	if err != nil {
		return fmt.Errorf("could not load a font face: %w", err)
	}

	p.ctx.SetFontFace(font)
	p.ctx.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 204})

	authorX := padding + float64(p.opts.AvaD) + padding/2
	authorY := padding + float64(p.opts.AvaD)/2

	p.ctx.DrawStringAnchored(p.opts.Author, authorX, authorY, 0, 0.5)

	return nil
}

func (p *Preview) drawTitle() error {
	font, err := loadFont(p.opts.TitleSize)

	if err != nil {
		return fmt.Errorf("could not load a font face: %w", err)
	}

	p.ctx.SetFontFace(font)
	p.ctx.SetColor(color.White)

	titleX := padding
	titleY := padding*2 + float64(p.opts.AvaD)
	maxWidth := float64(p.opts.CanvasW) - padding - margin*2
	title := p.opts.Title

	if utf8.RuneCountInString(title) > maxTitleLength {
		title = string([]rune(title)[0:maxTitleLength]) + "…"
	}

	p.ctx.DrawStringWrapped(title, titleX, titleY, 0, 0, maxWidth, 1.2, gg.AlignLeft)

	return nil
}

func (p *Preview) drawLogo(logoBuf []byte) error {
	logoBuf, err := scale(logoBuf, p.opts.LogoH)

	if err != nil {
		return fmt.Errorf("could not resize the logo: %w", err)
	}

	logoImg, _, err := image.Decode(bytes.NewReader(logoBuf))

	if err != nil {
		return fmt.Errorf("could not decode the logo: %w", err)
	}

	logoX := p.opts.CanvasW - padding - logoImg.Bounds().Dx()
	logoY := p.opts.CanvasH - padding - p.opts.LogoH

	p.ctx.DrawImage(logoImg, logoX, logoY)

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

// scale resizes an image to the specified height if it differs. Width of the image is auto.
func scale(buf []byte, h int) ([]byte, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(buf))

	if err != nil {
		return nil, err
	}

	if config.Height == h {
		return buf, nil
	}

	log.Printf("Scaling an image to %dpx height", h)

	vipsImg, err := vips.NewImageFromBuffer(buf)

	if err != nil {
		return nil, err
	}

	defer vipsImg.Close()

	if err = vipsImg.Resize(float64(h)/float64(config.Height), vips.KernelAuto); err != nil {
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

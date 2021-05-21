package drawer

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
	maxTitleLength = 90
)

// Options defines a set of options required to draw a preview.
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

// Draw draws a preview using the options provided.
func Draw(opts Options) (image.Image, error) {
	preview := gg.NewContext(opts.CanvasW, opts.CanvasH)

	// draw the background
	bgBuf, err := os.ReadFile(opts.BgURL)

	if err != nil {
		return nil, fmt.Errorf("could not get the background %s: %w", opts.BgURL, err)
	}

	bgBuf, err = resize(bgBuf, opts.CanvasW, opts.CanvasH)

	if err != nil {
		return nil, fmt.Errorf("could not resize the background: %w", err)
	}

	bgImg, _, err := image.Decode(bytes.NewReader(bgBuf))

	if err != nil {
		return nil, fmt.Errorf("could not decode the background: %w", err)
	}

	preview.DrawImage(bgImg, 0, 0)

	// draw the semi-transparent layer
	preview.SetColor(color.RGBA{0, 0, 0, 155})
	preview.DrawRectangle(margin, margin, float64(opts.CanvasW)-(margin*2), float64(opts.CanvasH)-(margin*2))
	preview.Fill()

	// draw the avatar border circle
	border := 8
	avaX := float64(padding + (opts.AvaD+border)/2)
	avaY := float64(padding + (opts.AvaD+border)/2)

	preview.DrawCircle(avaX, avaY, float64((opts.AvaD+8)/2))
	preview.SetHexColor("#FFFFFF")
	preview.Fill()

	// draw the avatar itself (cropped to a circle)
	avaBuf, err := os.ReadFile(opts.AvaURL)

	if err != nil {
		return nil, fmt.Errorf("could not get the avatar %s: %w", opts.AvaURL, err)
	}

	avaBuf, err = resize(avaBuf, opts.AvaD, opts.AvaD)

	if err != nil {
		return nil, fmt.Errorf("could not resize the avatar: %w", err)
	}

	avaImg, _, err := image.Decode(bytes.NewReader(avaBuf))

	if err != nil {
		return nil, fmt.Errorf("could not decode the avatar: %w", err)
	}

	avaImg = circle(avaImg)

	preview.DrawImageAnchored(avaImg, int(avaX), int(avaY), 0.5, 0.5)

	// draw the author name
	preview.LoadFontFace(filepath.Join("fonts", "Ubuntu-Medium.ttf"), 36)
	preview.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 204})

	authorX := avaX + float64(opts.AvaD)/2 + padding/2
	authorY := float64(opts.AvaD)/2 + padding

	preview.DrawStringAnchored(opts.Author, authorX, authorY, 0, 0.5)

	// draw the title
	preview.LoadFontFace(filepath.Join("fonts", "Ubuntu-Medium.ttf"), 76)
	preview.SetColor(color.White)

	titleX := padding
	titleY := authorY + float64(opts.AvaD)/2 + padding
	maxWidth := float64(opts.CanvasW) - padding - margin*2
	title := opts.Title

	if utf8.RuneCountInString(title) > maxTitleLength {
		title = string([]rune(title)[0:maxTitleLength]) + "…"
	}

	preview.DrawStringWrapped(title, titleX, titleY, 0, 0, maxWidth, 1.6, gg.AlignLeft)

	// draw the required right part of the label
	preview.LoadFontFace(filepath.Join("fonts", "Ubuntu-Bold.ttf"), 36)
	preview.SetColor(color.White)

	labelRightWidth, labelRightHeight := preview.MeasureString(opts.LabelR)
	labelX := float64(opts.CanvasW) - labelRightWidth - padding
	labelY := float64(opts.CanvasH) - padding

	preview.DrawString(opts.LabelR, labelX, labelY)

	// draw the icon
	iconBuf, err := os.ReadFile(opts.IconURL)

	if err != nil {
		return nil, fmt.Errorf("could not get the icon: %w", err)
	}

	iconBuf, err = resize(iconBuf, opts.IconW, opts.IconH)

	if err != nil {
		return nil, fmt.Errorf("could not resize the icon: %w", err)
	}

	iconImg, _, err := image.Decode(bytes.NewReader(iconBuf))

	if err != nil {
		return nil, fmt.Errorf("could not decode the icon: %w", err)
	}

	iconX := int(labelX) - opts.IconW - int(margin/2)
	iconY := opts.CanvasH - int(padding) - int(labelRightHeight/2)

	preview.DrawImageAnchored(iconImg, iconX, iconY, 0, 0.5)

	if len(opts.LabelL) > 0 {
		labelLeftWidth, _ := preview.MeasureString(opts.LabelL)

		labelX = float64(opts.CanvasW) - labelLeftWidth - labelRightWidth - float64(opts.IconW) - margin - padding
		labelY = float64(opts.CanvasH) - padding

		preview.DrawString(opts.LabelL, labelX, labelY)
	}

	return preview.Image(), nil
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

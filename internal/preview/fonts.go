package preview

import (
	"embed"
	"sync"

	"github.com/AndreKR/multiface"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

const (
	textFont    = "fonts/Ubuntu-Medium.ttf"
	symbolsFont = "fonts/NotoSansSymbols-Medium.ttf"
	emoji1Font  = "fonts/NotoEmoji-Regular.ttf"
	emoji2Font  = "fonts/Symbola.ttf"
)

//go:embed fonts/*
var fonts embed.FS
var cache sync.Map

// loadFont loads a multiface consisting of letters, symbols and emojis merged to one font face.
// It caches the result in memory for each font size to avoid multiface creation on each request
func loadFont(points float64) (font.Face, error) {
	if cached, exists := cache.Load(points); exists {
		if face, ok := cached.(font.Face); ok {
			return face, nil
		}
	}

	face := new(multiface.Face)
	textBuf, err := fonts.ReadFile(textFont)

	if err != nil {
		return nil, err
	}

	textFont, err := truetype.Parse(textBuf)

	if err != nil {
		return nil, err
	}

	textFace := truetype.NewFace(textFont, &truetype.Options{
		Size: points,
	})

	face.AddTruetypeFace(textFace, textFont)

	symbolsBuf, err := fonts.ReadFile(symbolsFont)

	if err != nil {
		return nil, err
	}

	symbolsFont, err := truetype.Parse(symbolsBuf)

	if err != nil {
		return nil, err
	}

	symbolsFace := truetype.NewFace(symbolsFont, &truetype.Options{
		Size: points,
	})

	face.AddTruetypeFace(symbolsFace, symbolsFont)

	emoji1Buf, err := fonts.ReadFile(emoji1Font)

	if err != nil {
		return nil, err
	}

	emoji1Font, err := truetype.Parse(emoji1Buf)

	if err != nil {
		return nil, err
	}

	emoji1Face := truetype.NewFace(emoji1Font, &truetype.Options{
		Size: points,
	})

	face.AddTruetypeFace(emoji1Face, emoji1Font)

	emoji2Buf, err := fonts.ReadFile(emoji2Font)

	if err != nil {
		return nil, err
	}

	emoji2Font, err := truetype.Parse(emoji2Buf)

	if err != nil {
		return nil, err
	}

	emoji2Face := truetype.NewFace(emoji2Font, &truetype.Options{
		Size: points,
	})

	face.AddTruetypeFace(emoji2Face, emoji2Font)
	cache.Store(points, face)

	return face, nil
}

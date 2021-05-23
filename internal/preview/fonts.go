package preview

import (
	"embed"
	"sync"

	"github.com/AndreKR/multiface"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
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

	emojiBuf, err := fonts.ReadFile(emojiFont)

	if err != nil {
		return nil, err
	}

	emojiFont, err := truetype.Parse(emojiBuf)

	if err != nil {
		return nil, err
	}

	emojiFace := truetype.NewFace(emojiFont, &truetype.Options{
		Size: points,
	})

	face.AddTruetypeFace(emojiFace, emojiFont)
	cache.Store(points, face)

	return face, nil
}

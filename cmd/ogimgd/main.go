package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/nDmitry/ogimgd/internal/preview"
)

type drawer interface {
	Draw() (image.Image, error)
}

func main() {
	vips.LoggingSettings(nil, vips.LogLevelError)

	vips.Startup(nil)
	defer vips.Shutdown()

	p := preview.New(preview.Options{
		CanvasW: 1200,
		CanvasH: 630,
		Opacity: 0.6,
		AvaD:    64,
		Title:   "Инвестиции в зарубежные бумаги через российских брокеров: варианты и подводные камни",
		Author:  "@DmitryNikitenko",
		LabelL:  "Rational",
		LabelR:  "Answer",
		BgURL:   "https://images.unsplash.com/photo-1534469589579-86bd01bc003a",
		AvaURL:  "https://avatars.githubusercontent.com/u/2134568?v=4",
		IconURL: "https://i.imgur.com/PqhZXkZ.png",
		IconW:   48,
		IconH:   48,
	})

	if err := run(p); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(d drawer) error {
	img, err := d.Draw()

	if err != nil {
		return err
	}

	f, err := os.Create("out.jpg")

	if err != nil {
		panic(err)
	}

	defer f.Close()
	jpeg.Encode(f, img, &jpeg.Options{Quality: 84})

	return nil
}

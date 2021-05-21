package main

import (
	"fmt"
	"image/jpeg"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/nDmitry/previewer/internal/drawer"
)

func main() {
	vips.LoggingSettings(nil, vips.LogLevelError)

	vips.Startup(nil)
	defer vips.Shutdown()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	img, err := drawer.Draw(drawer.Options{
		CanvasW: 1200,
		CanvasH: 630,
		AvaD:    64,
		Title:   "Инвестиции в зарубежные бумаги через российских брокеров: варианты и подводные камни",
		Author:  "@DmitryNikitenko",
		LabelL:  "Rational",
		LabelR:  "Answer",
		BgURL:   "./bg.jpg",
		AvaURL:  "./avatar.jpg",
		IconURL: "./icon.png",
		IconW:   48,
		IconH:   48,
	})

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

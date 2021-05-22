package main

import (
	"log"
	"os"
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/nDmitry/ogimgd/internal/preview"
	"github.com/nDmitry/ogimgd/internal/server"
)

func main() {
	port := 8201

	if os.Getenv("PORT") != "" {
		var err error

		if port, err = strconv.Atoi(os.Getenv("PORT")); err != nil {
			log.Fatalf("could not parse the app port: %s\n", os.Getenv("PORT"))
		}
	}

	vips.LoggingSettings(nil, vips.LogLevelError)

	vips.Startup(nil)
	defer vips.Shutdown()

	p := preview.New()

	server.Run(port, p)
}

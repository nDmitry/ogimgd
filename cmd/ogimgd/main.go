package main

import (
	"flag"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/nDmitry/ogimgd/internal/preview"
	"github.com/nDmitry/ogimgd/internal/server"
)

func main() {
	var port = *flag.Int("port", 8201, "HTTP server port to listen")

	vips.LoggingSettings(nil, vips.LogLevelError)

	vips.Startup(nil)
	defer vips.Shutdown()

	p := preview.New()

	server.Run(port, p)
}

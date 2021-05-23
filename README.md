# ogimgd

Social previews generator as a microservice. Can be used to generate images for `og:image` meta-tag.

It runs as an HTTP server with a single endpoint `/preview` that accepts various query parameters to customize the output preview image:

* `title` (string, required) - text you'd like do display on the image (90 characters max, the rest will be trimmed and replaced with …).
* `author` (string, required) - a user name or handle to display above the `title`
* `ava` (string, required) - a URL to a remote user avatar image that will be downloaded via HTTP and placed beside the `author` name.
* `logo` (string, required) - a URL to a remote image that will be placed at the bottom right corner of the preview.
* `bg` (string, optional) - a URL to a remote image that will be used as a background of the preview. Or a HEX-color (starting with #, e.g. `#FFA` or `#FFFAAA`) in case the image is missing or you prefer a blank color.
* `op` (float, optional, default 0.6) - opacity value for the black foreground under the text elements of the preview.

Wherever a URL is expected, you can also pass a filename to a local image located in the `internal/remote/images` folder. It can be used with images that don't change (e.g. logo) to save some network roundtrips.

If you control remote images sizes, you can check the default sizes in [options](https://github.com/nDmitry/ogimgd/blob/main/internal/server/handlers.go#L29) and prepare images in advance to avoid resizing.

See the example requests in [requests.http](https://github.com/nDmitry/ogimgd/blob/main/requests.http) file.

## Preview example

![Preview example](/internal/server/testdata/expected/bg-remote.jpeg?raw=true)

Search for more in [tests expected images](https://github.com/nDmitry/ogimgd/blob/main/internal/server/testdata/expected/).

## Running

`make up` will spin up a server in a Docker container. By default it will listen on the port 8201 that can be changed using `PORT` environment variable.
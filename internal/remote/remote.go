package remote

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

const bodyLimit = 10 * 1024 * 1024

//go:embed images/*
var images embed.FS

// Remote can obtain remote resources to use in the preview.
type Remote struct {
	httpClient *http.Client
}

// New returns an initialized Remote.
func New() *Remote {
	return &Remote{
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}
}

// Get fetches a remote resource using an URL or try to read it from the disk when a filename is specified.
func (r *Remote) Get(ctx context.Context, urlOrPath string) (buf []byte, err error) {
	log.Printf("getting a resource: %s\n", urlOrPath)

	_, parseErr := url.ParseRequestURI(urlOrPath)

	// expects a filename if it doesn't look like an URL
	if parseErr != nil {
		// replace here is paranoia (base path extraction is already enough)
		filename := filepath.Base(strings.ReplaceAll(urlOrPath, "../", ""))
		buf, err = images.ReadFile(filepath.Join("images", filename))

		if err != nil {
			return nil, fmt.Errorf("could not parse a URL nor open a file with this filename: %s: %w", filename, err)
		}

		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlOrPath, nil)

	if err != nil {
		return nil, fmt.Errorf("could not get a resource by the url: %s: %w", urlOrPath, err)
	}

	res, err := r.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("could not get a resource by the url: %s: %w", urlOrPath, err)
	}

	defer res.Body.Close()

	buf, err = ioutil.ReadAll(io.LimitReader(res.Body, bodyLimit))

	if err != nil {
		return nil, fmt.Errorf("could not read a resource body: %s: %w", urlOrPath, err)
	}

	return
}

// GetAll fetches remote resources concurrently using Get
func (r *Remote) GetAll(ctx context.Context, urlsOrPaths map[string]string) (map[string][]byte, error) {
	bufs := make(map[string][]byte, len(urlsOrPaths))
	errCh := make(chan error)
	doneCh := make(chan bool)
	var wg sync.WaitGroup

	wg.Add(len(urlsOrPaths))

	for key, urlOrPath := range urlsOrPaths {
		go func(key string, urlOrPath string) {
			buf, err := r.Get(ctx, urlOrPath)

			if err != nil {
				errCh <- err
			}

			bufs[key] = buf

			wg.Done()
		}(key, urlOrPath)
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		close(errCh)

		return bufs, nil
	case err := <-errCh:
		return nil, err
	}
}

package remote

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Remote struct {
	httpClient *http.Client
}

func New() *Remote {
	return &Remote{
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}
}

func (r *Remote) Get(url string) ([]byte, error) {
	res, err := r.httpClient.Get(url)

	if err != nil {
		return nil, fmt.Errorf("could not get a resource by the url: %s: %w", url, err)
	}

	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("could not read a resource body: %s: %w", url, err)
	}

	return buf, nil
}

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/nDmitry/ogimgd/internal/preview"
)

func TestPreviewHandler_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile("./testdata/bg.jpg")

		if err != nil {
			t.Error(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(file)
	}))

	defer ts.Close()

	p := preview.New()
	handler := getPreview(p)

	testCases := []struct {
		name     string
		req      string
		expected string
	}{{
		name:     "basic",
		req:      "/preview?title=The%20quick%20brown%20fox%20jumps%20over%20the%20lazy%20dog&author=%40Tester&ava=avatar.png&logo=logo.png",
		expected: "./testdata/expected/basic.jpeg",
	}, {
		name:     "opacity",
		req:      "/preview?title=The%20quick%20brown%20fox%20jumps%20over%20the%20lazy%20dog&author=%40Tester&ava=avatar.png&logo=logo.png&op=0.5",
		expected: "./testdata/expected/opacity.jpeg",
	}, {
		name:     "bg color",
		req:      "/preview?title=The%20quick%20brown%20fox%20jumps%20over%20the%20lazy%20dog&author=%40Tester&ava=avatar.png&logo=logo.png&bg=%23FFA",
		expected: "./testdata/expected/bg-color.jpeg",
	}, {
		name:     "bg remote",
		req:      fmt.Sprintf("/preview?title=%s&author=%s&ava=avatar.png&logo=logo.png&bg=%s", url.QueryEscape("The quick brown fox jumps over the lazy dog"), url.QueryEscape("@Tester"), url.QueryEscape(ts.URL)),
		expected: "./testdata/expected/bg-remote.jpeg",
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				"GET",
				tt.req,
				nil,
			)

			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			body, err := ioutil.ReadAll(res.Body)
			expected, err := os.ReadFile(tt.expected)

			if err != nil {
				t.Error(err)
			}

			if bytes.Compare(body, expected) != 0 {
				t.Error("images are not equal")
			}
		})
	}
}

func TestPreviewHandler_Bad(t *testing.T) {
	p := preview.New()
	handler := getPreview(p)

	testCases := []struct {
		name     string
		req      string
		expected string
	}{{
		name:     "title",
		req:      "/preview?author=%40Tester&ava=avatar.png&logo=logo.png",
		expected: "Missing required title parameter",
	}, {
		name:     "author",
		req:      "/preview?title=The%20quick%20brown%20fox%20jumps%20over%20the%20lazy%20dog&ava=avatar.png&logo=logo.png",
		expected: "Missing required author parameter",
	}, {
		name:     "ava",
		req:      "/preview?title=The%20quick%20brown%20fox%20jumps%20over%20the%20lazy%20dog&author=%40Tester&logo=logo.png",
		expected: "Missing required ava parameter",
	}, {
		name:     "logo",
		req:      "/preview?title=The%20quick%20brown%20fox%20jumps%20over%20the%20lazy%20dog&author=%40Tester&ava=avatar.png",
		expected: "Missing required logo parameter",
	}, {
		name:     "opacity",
		req:      "/preview?title=The%20quick%20brown%20fox%20jumps%20over%20the%20lazy%20dog&author=%40Tester&ava=avatar.png&logo=logo.png&op=bad",
		expected: "Could not parse op parameter",
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				"GET",
				tt.req,
				nil,
			)

			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			body, err := ioutil.ReadAll(res.Body)

			if err != nil {
				t.Error(err)
			}

			mes := errorResponse{}
			err = json.Unmarshal(body, &mes)

			if err != nil {
				t.Error(err)
			}

			if mes.Message != tt.expected {
				t.Errorf("error messages are not equal, expected: %s, actual: %s", tt.expected, mes.Message)
			}
		})
	}
}

package reactions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	original := baseURL
	baseURL = server.URL
	t.Cleanup(func() { baseURL = original })
}

func TestFetchImageURL_Success(t *testing.T) {
	var gotPath string
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(waifuPicsResponse{URL: "https://example.com/image.gif"})
	})

	url, ok, err := FetchImageURL(context.Background(), "hug", false)
	if err != nil {
		t.Fatalf("FetchImageURL() error = %v", err)
	}
	if !ok {
		t.Fatal("FetchImageURL() ok = false, want true")
	}
	if url != "https://example.com/image.gif" {
		t.Fatalf("url = %q, want %q", url, "https://example.com/image.gif")
	}
	if gotPath != "/sfw/hug" {
		t.Fatalf("requested path = %q, want %q", gotPath, "/sfw/hug")
	}
}

func TestFetchImageURL_NSFWPath(t *testing.T) {
	var gotPath string
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(waifuPicsResponse{URL: "https://example.com/x.gif"})
	})

	if _, _, err := FetchImageURL(context.Background(), "waifu", true); err != nil {
		t.Fatalf("FetchImageURL() error = %v", err)
	}
	if gotPath != "/nsfw/waifu" {
		t.Fatalf("requested path = %q, want %q", gotPath, "/nsfw/waifu")
	}
}

func TestFetchImageURL_NonOKStatus(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, ok, err := FetchImageURL(context.Background(), "hug", false)
	if err != nil {
		t.Fatalf("FetchImageURL() error = %v, want nil (a bad status is a soft failure)", err)
	}
	if ok {
		t.Fatal("FetchImageURL() ok = true, want false on a non-200 response")
	}
}

func TestFetchImageURL_MalformedBody(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})

	_, ok, err := FetchImageURL(context.Background(), "hug", false)
	if err != nil {
		t.Fatalf("FetchImageURL() error = %v, want nil", err)
	}
	if ok {
		t.Fatal("FetchImageURL() ok = true, want false on a malformed body")
	}
}

func TestFetchImageURL_EmptyURLInBody(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(waifuPicsResponse{URL: ""})
	})

	_, ok, err := FetchImageURL(context.Background(), "hug", false)
	if err != nil {
		t.Fatalf("FetchImageURL() error = %v, want nil", err)
	}
	if ok {
		t.Fatal("FetchImageURL() ok = true, want false when the response has no url")
	}
}

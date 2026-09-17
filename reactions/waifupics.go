// Package reactions implements the anime-reaction-image commands (/sfw,
// /reaction, /nsfw) ported from the original Python bot's reactions.py,
// backed by the public https://waifu.pics API.
package reactions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	httpClient = &http.Client{Timeout: 10 * time.Second}
	baseURL    = "https://api.waifu.pics" // overridden in tests
)

type waifuPicsResponse struct {
	URL string `json:"url"`
}

// FetchImageURL fetches a random image URL for the given waifu.pics
// category, from the sfw or nsfw endpoint. It returns ok=false (with no
// error) if the API responded but didn't include a usable URL, mirroring
// the original bot's "couldn't reach the image database" fallback path.
func FetchImageURL(ctx context.Context, category string, nsfw bool) (url string, ok bool, err error) {
	kind := "sfw"
	if nsfw {
		kind = "nsfw"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s/%s", baseURL, kind, category), nil)
	if err != nil {
		return "", false, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, nil
	}

	var body waifuPicsResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", false, nil //nolint:nilerr // a malformed response is "couldn't get an image", not a hard error
	}
	if body.URL == "" {
		return "", false, nil
	}
	return body.URL, true, nil
}

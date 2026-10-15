package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	thumbBase   = "https://thumbnails.roblox.com"
	userAgent   = "danh-birthday-fetch-assets/1.0"
	maxAttempts = 3
)

var client = &http.Client{Timeout: 15 * time.Second}

type thumbResponse struct {
	Data []struct {
		TargetID int64  `json:"targetId"`
		State    string `json:"state"`
		ImageURL string `json:"imageUrl"`
	} `json:"data"`
}

func getJSON(url string, out any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// thumbURLs resolves a thumbnails API endpoint to image URLs keyed by targetId.
// Thumbnails can be "Pending" while Roblox renders them, so it retries.
func thumbURLs(endpoint string) (map[int64]string, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var tr thumbResponse
		if err := getJSON(thumbBase+endpoint, &tr); err != nil {
			lastErr = err
		} else {
			urls := map[int64]string{}
			pending := false
			for _, d := range tr.Data {
				switch {
				case d.State == "Completed" && d.ImageURL != "":
					urls[d.TargetID] = d.ImageURL
				case d.State == "Pending":
					pending = true
				}
			}
			if !pending && len(urls) > 0 {
				return urls, nil
			}
			lastErr = fmt.Errorf("thumbnail not ready (%d ready, pending=%v)", len(urls), pending)
		}
		time.Sleep(2 * time.Second)
	}
	return nil, lastErr
}

// thumbURL is thumbURLs for endpoints that return a single image.
func thumbURL(endpoint string) (string, error) {
	urls, err := thumbURLs(endpoint)
	if err != nil {
		return "", err
	}
	for _, u := range urls {
		return u, nil
	}
	return "", fmt.Errorf("no image for %s", endpoint)
}

// currentlyWearing returns asset IDs the user is wearing, in API order.
func currentlyWearing(userID int64) ([]int64, error) {
	var out struct {
		AssetIDs []int64 `json:"assetIds"`
	}
	url := fmt.Sprintf("https://avatar.roblox.com/v1/users/%d/currently-wearing", userID)
	if err := getJSON(url, &out); err != nil {
		return nil, err
	}
	return out.AssetIDs, nil
}

// gameThumbURL resolves the first screenshot of a universe. The multiget
// endpoint nests images one level deeper than the other thumbnail endpoints.
func gameThumbURL(universeID int64) (string, error) {
	var out struct {
		Data []struct {
			Thumbnails []struct {
				State    string `json:"state"`
				ImageURL string `json:"imageUrl"`
			} `json:"thumbnails"`
		} `json:"data"`
	}
	url := fmt.Sprintf("%s/v1/games/multiget/thumbnails?universeIds=%d&size=768x432&format=Png&count=1", thumbBase, universeID)
	if err := getJSON(url, &out); err != nil {
		return "", err
	}
	for _, d := range out.Data {
		for _, t := range d.Thumbnails {
			if t.State == "Completed" && t.ImageURL != "" {
				return t.ImageURL, nil
			}
		}
	}
	return "", fmt.Errorf("no completed thumbnail for universe %d", universeID)
}

// assetDetails looks up a catalog item's display name and type; zero values on failure.
func assetDetails(id int64) (name string, typeID int) {
	var out struct {
		Name        string `json:"Name"`
		AssetTypeID int    `json:"AssetTypeId"`
	}
	if err := getJSON(fmt.Sprintf("https://economy.roblox.com/v2/assets/%d/details", id), &out); err != nil {
		return "", 0
	}
	return out.Name, out.AssetTypeID
}

// isShowcaseType reports whether an asset type is worth showing on the
// profile card: animations (48–56, 61) and body parts (17, 18, 27–31) are not.
func isShowcaseType(t int) bool {
	switch {
	case t == 0, t == 17, t == 18, t >= 27 && t <= 31, t >= 48 && t <= 56, t == 61:
		return false
	}
	return true
}

func joinIDs(ids []int64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}

// Command fetch-assets downloads every remote image the birthday site uses
// into site/assets so the page never hotlinks expiring Roblox CDN URLs.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
)

const (
	danhUserID         = 1037207842
	bloxFruitsUniverse = 994732206
	stealAnEggUniverse = 10563114921
	fruitCDN           = "https://fruityblox-storage.b-cdn.net/"
	maxWearingItems    = 4
)

// pageBG matches --bg in site/css/tokens.css (JPEG flatten color).
var pageBG = color.RGBA{R: 0x0b, G: 0x10, B: 0x26, A: 0xff}

type fruit struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Rarity string `json:"rarity"`
	Img    string `json:"img"`
}

var fruits = []fruit{
	{ID: "dragon-west", Name: "Dragon", Rarity: "Mythical"},
	{ID: "kitsune", Name: "Kitsune", Rarity: "Mythical"},
	{ID: "dough", Name: "Dough", Rarity: "Mythical"},
	{ID: "venom", Name: "Venom", Rarity: "Mythical"},
	{ID: "spirit", Name: "Spirit", Rarity: "Mythical"},
	{ID: "t-rex", Name: "T-Rex", Rarity: "Mythical"},
	{ID: "buddha", Name: "Buddha", Rarity: "Legendary"},
	{ID: "light", Name: "Light", Rarity: "Rare"},
	{ID: "flame", Name: "Flame", Rarity: "Uncommon"},
	{ID: "ice", Name: "Ice", Rarity: "Uncommon"},
}

type wearingItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Img  string `json:"img"`
}

type runner struct {
	out    string
	force  bool
	failed int
}

// fetch resolves (if needed) and downloads one asset, printing a status line.
// rel is relative to the assets dir; resolve may be nil for static URLs.
func (r *runner) fetch(rel string, resolve func() (string, error), url string) bool {
	dest := filepath.Join(r.out, filepath.FromSlash(rel))
	if !r.force {
		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("  skip  %s\n", rel)
			return true
		}
	}
	if resolve != nil {
		u, err := resolve()
		if err != nil {
			return r.fail(rel, err)
		}
		url = u
	}
	if err := download(url, dest, r.force); err != nil && !errors.Is(err, errExists) {
		return r.fail(rel, err)
	}
	fmt.Printf("  ok    %s\n", rel)
	return true
}

// fetchImage downloads a PNG next to rel, then shrinks it into rel per spec.
func (r *runner) fetchImage(rel string, resolve func() (string, error), spec optimizeSpec) bool {
	dest := filepath.Join(r.out, filepath.FromSlash(rel))
	if !r.force {
		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("  skip  %s\n", rel)
			return true
		}
	}
	rawRel := rel + ".src.png"
	if !r.fetch(rawRel, resolve, "") {
		return false
	}
	if err := optimize(filepath.Join(r.out, filepath.FromSlash(rawRel)), dest, spec); err != nil {
		return r.fail(rel, err)
	}
	if st, err := os.Stat(dest); err == nil {
		fmt.Printf("  opt   %s (%d KB)\n", rel, st.Size()/1024)
	}
	return true
}

func (r *runner) fail(rel string, err error) bool {
	r.failed++
	fmt.Printf("  FAIL  %s: %v\n", rel, err)
	return false
}

func thumb(endpoint string) func() (string, error) {
	return func() (string, error) { return thumbURL(endpoint) }
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func main() {
	out := flag.String("out", "../../site/assets", "assets output directory")
	force := flag.Bool("force", false, "re-download files that already exist")
	flag.Parse()
	r := &runner{out: *out, force: *force}

	fmt.Println("Danh's avatar")
	okAvatar := r.fetchImage("img/avatar.png", thumb(fmt.Sprintf("/v1/users/avatar?userIds=%d&size=720x720&format=Png&isCircular=false", danhUserID)), optimizeSpec{maxW: 520})
	okHead := r.fetchImage("img/headshot.png", thumb(fmt.Sprintf("/v1/users/avatar-headshot?userIds=%d&size=420x420&format=Png&isCircular=false", danhUserID)), optimizeSpec{maxW: 180})

	fmt.Println("Game art")
	for name, u := range map[string]int64{"blox-fruits": bloxFruitsUniverse, "steal-an-egg": stealAnEggUniverse} {
		r.fetchImage("img/"+name+"-icon.jpg", thumb(fmt.Sprintf("/v1/games/icons?universeIds=%d&size=256x256&format=Png&isCircular=false", u)), optimizeSpec{maxW: 144, jpeg: true, quality: 82})
		r.fetchImage("img/"+name+"-thumb.jpg", func() (string, error) { return gameThumbURL(u) }, optimizeSpec{jpeg: true, quality: 68})
	}

	fmt.Println("Blox Fruits")
	var gotFruits []fruit
	for _, f := range fruits {
		f.Img = "assets/img/fruits/" + f.ID + ".webp"
		if r.fetch("img/fruits/"+f.ID+".webp", nil, fruitCDN+f.ID+".webp") {
			gotFruits = append(gotFruits, f)
		}
	}
	if err := writeJSON(filepath.Join(r.out, "data", "fruits.json"), gotFruits); err != nil {
		r.fail("data/fruits.json", err)
	}

	fmt.Println("Currently wearing")
	if err := writeJSON(filepath.Join(r.out, "data", "wearing.json"), r.fetchWearing()); err != nil {
		r.fail("data/wearing.json", err)
	}

	fmt.Printf("done: %d failure(s)\n", r.failed)
	if !okAvatar || !okHead {
		os.Exit(1) // the hero section cannot work without these two
	}
}

// fetchWearing is best-effort: any error just yields fewer items.
func (r *runner) fetchWearing() []wearingItem {
	items := []wearingItem{}
	ids, err := currentlyWearing(danhUserID)
	if err != nil {
		r.fail("currently-wearing", err)
		return items
	}
	if len(ids) == 0 {
		return items
	}
	urls, err := thumbURLs(fmt.Sprintf("/v1/assets?assetIds=%s&size=150x150&format=Png&isCircular=false", joinIDs(ids)))
	if err != nil {
		r.fail("wearing thumbnails", err)
		return items
	}
	// Newest items sit at the end of the list and are the interesting ones
	// (hair, accessories, clothing), so walk it backwards.
	for i := len(ids) - 1; i >= 0 && len(items) < maxWearingItems; i-- {
		id := ids[i]
		u, ok := urls[id]
		if !ok {
			continue
		}
		name, typeID := assetDetails(id)
		if !isShowcaseType(typeID) {
			continue
		}
		rel := fmt.Sprintf("img/wearing/%d.png", id)
		if !r.fetch(rel, nil, u) {
			continue
		}
		items = append(items, wearingItem{ID: id, Name: name, Img: "assets/" + rel})
	}
	return items
}

package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

// optimizeSpec shrinks a downloaded PNG so the page stays light on mobile.
// JPEG drops transparency, so it is only used for opaque game art.
type optimizeSpec struct {
	maxW    int  // downscale to this width if wider (0 = keep size)
	jpeg    bool // re-encode as JPEG instead of PNG
	quality int  // JPEG quality
}

// optimize reads src (PNG), resizes/re-encodes it per spec, writes dest and
// removes src when the two differ.
func optimize(src, dest string, spec optimizeSpec) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(src), err)
	}

	if b := img.Bounds(); spec.maxW > 0 && b.Dx() > spec.maxW {
		h := b.Dy() * spec.maxW / b.Dx()
		dst := image.NewNRGBA(image.Rect(0, 0, spec.maxW, h))
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Src, nil)
		img = dst
	}

	tmp := dest + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if spec.jpeg {
		err = jpeg.Encode(out, flatten(img), &jpeg.Options{Quality: spec.quality})
	} else {
		err = (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(out, img)
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}
	if src != dest {
		return os.Remove(src)
	}
	return nil
}

// flatten paints img over the page background so transparent corners don't turn black in JPEG.
func flatten(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(pageBG), image.Point{}, draw.Src)
	draw.Draw(dst, b, img, b.Min, draw.Over)
	return dst
}

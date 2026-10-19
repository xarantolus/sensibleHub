package music

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeImage(t *testing.T, name string, width int, encode func(*bytes.Buffer, image.Image) error) (string, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if err := encode(&buf, image.NewRGBA(image.Rect(0, 0, width, width))); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p, buf.Bytes()
}

func encodeJPEG(b *bytes.Buffer, img image.Image) error { return jpeg.Encode(b, img, nil) }
func encodePNG(b *bytes.Buffer, img image.Image) error  { return png.Encode(b, img) }

func TestCoverJPEG(t *testing.T) {
	small, smallBytes := writeImage(t, "small.jpg", 40, encodeJPEG)
	large, _ := writeImage(t, "large.jpg", 120, encodeJPEG)
	pngCover, _ := writeImage(t, "cover.png", 40, encodePNG)

	got, err := coverJPEG(small, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, smallBytes) {
		t.Error("a JPEG within the size limit must be embedded unchanged")
	}

	for _, tc := range []struct {
		path      string
		maxSize   int
		wantWidth int
	}{
		{large, 100, 100},
		{large, 0, 120},
		{pngCover, 100, 40},
	} {
		got, err := coverJPEG(tc.path, tc.maxSize)
		if err != nil {
			t.Fatal(err)
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(got))
		if err != nil {
			t.Fatal(err)
		}
		if format != "jpeg" || cfg.Width != tc.wantWidth {
			t.Errorf("%s (max %d): got %s %dpx wide, want jpeg %dpx", filepath.Base(tc.path), tc.maxSize, format, cfg.Width, tc.wantWidth)
		}
	}
}

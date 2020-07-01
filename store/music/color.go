package music

import (
	"os"
	"image"
	"github.com/cenkalti/dominantcolor"
)

// CalculateDominantColor calculates the dominant color of the image at the given path
func CalculateDominantColor(coverPath string ) (hex string, err error) {
	f, err := os.Open(coverPath)
	if err != nil {
		return 
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return
	}

	return dominantcolor.Hex(dominantcolor.Find(img)), nil
}
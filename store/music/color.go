package music

import (
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"os"

	"github.com/cenkalti/dominantcolor"
)

// CalculateDominantColor calculates the dominant color of the image at the given path
func CalculateDominantColor(coverPath string) (hex string, err error) {
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

type Color string // e.g. "#000000"

// BorderColor returns the color for the border around this picture
func (c Color) BorderColor() string {
	col, err := parseHexColorFast(string(c))
	if err != nil {
		return string(c)
	}

	const limit = 224

	// If the color is white, we use a different border color as then the border cannot be seen
	if col.R > limit && col.G > limit && col.B > limit {
		// simply invert the color
		col = color.RGBA{
			255 - col.R,
			255 - col.G,
			255 - col.B,
			255,
		}

		// no need for col.A as it is 255
		return "#" + hex.EncodeToString([]byte{col.R, col.B, col.B})
	}

	return string(c)
}

var errInvalidFormat = errors.New("invalid format")

// https://stackoverflow.com/a/54200713/5728357
func parseHexColorFast(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = errInvalidFormat
	}
	return
}

package store

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"xarantolus/sensibleHub/store/file"
	"xarantolus/sensibleHub/store/music"

	"github.com/edwvee/exiffix"
	"github.com/vitali-fedulov/images"
)

// GenerateCoverPreviews starts generating all cover previews.
// It should be called once when the server starts up.
// Since it might take multiple minutes, it runs in the background.
// It is sequential, which means that it might be quite slow, but that usually
// doesn't matter when we just warm the cache and nobody is requesting anything
func (m *Manager) GenerateCoverPreviews() {
	if !m.cfg.GenerateOnStartup {
		return
	}

	newest, _ := m.Newest()
	if len(newest) == 0 {
		// We don't have any songs yet
		return
	}

	log.Println("[Startup]: Starting cover preview generation")

	// This might take some time, especially with more songs.
	// Could easily be parallelized, but I see no reason for that
	// Also, yes, the songs in newest are done twice, but again,
	// I don't really care.
	// The goal is to warm up the songs one sees first (main page),
	// then all others.
	for _, e := range append(newest, m.AllEntries()...) {
		_, _, _ = e.CoverPreview()
	}

	log.Printf("[Startup]: Finished cover preview generation, currently holding %d covers with a size of %d bytes\n", music.PreviewLen(), music.PreviewSize())
}

// cropMoveCover tries to create a squared cover image from the image located at `sourceFile`.
// If no squared image can be generated, no image will be generated.
func cropMoveCover(sourceFile, destination string) (err error) {
	f, err := os.Open(sourceFile)
	if err != nil {
		return
	}
	defer f.Close()

	return cropCover(f, sourceFile, destination)
}

// cropCover crops a cover image stored in `f` to a square and writes it to a file at `destination`.
// sourceFile can be "" (empty string) if the cover has been read e.g. from an http request
func cropCover(f io.Reader, sourceFile string, destination string) (err error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	// Since we need an io.ReadSeeker, we need to know all bytes
	img, format, err := exiffix.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	bounds := img.Bounds()

	var croppedImg image.Image

	// If we already have a square, we can just use the source file
	if bounds.Max.X == bounds.Max.Y {
		// Now we have a square - but is it the same file type as the one for the desired extension?
		desiredExtension := strings.ToUpper(filepath.Ext(destination))

		if (desiredExtension == ".JPG" || desiredExtension == ".JPEG") && (format == "jpeg" || format == "jpg") || desiredExtension == ".PNG" && format == "png" {
			// If yes, we don't need to worry about anything
			if sourceFile == "" {
				return ioutil.WriteFile(destination, data, 0o644)
			}

			return file.Move(sourceFile, destination)
		}

		croppedImg = img // work with the normal image
		goto noNeedToCrop
	}

	{
		// Use the smaller dimension for cutting off
		smallerOne := bounds.Max.X
		if bounds.Max.Y < smallerOne {
			smallerOne = bounds.Max.Y
		}

		// Basically take the middle square. This works e.g. with youtube music video thumbnails
		var defaultCrop image.Rectangle

		if bounds.Max.X > smallerOne {
			// Width is larger than height
			defaultCrop = image.Rect(bounds.Max.X/2-smallerOne/2, 0, bounds.Max.X/2+smallerOne/2, bounds.Max.Y)
		} else {
			// Height ist larger than width
			defaultCrop = image.Rect(0, bounds.Max.Y/2-smallerOne/2, bounds.Max.X, bounds.Max.Y/2+smallerOne/2)
		}

		type SubImager interface {
			SubImage(r image.Rectangle) image.Image
		}
		subImg, ok := img.(SubImager)

		// if we cannot crop, we won't use an image at all
		if !ok {
			return fmt.Errorf("cannot crop image")
		}

		croppedImg = subImg.SubImage(defaultCrop)
	}

noNeedToCrop:
	fn, err := encodeImageToTemp(croppedImg, destination)
	if err != nil {
		return
	}

	return file.Move(fn, destination)
}

func encodeImageToTemp(img image.Image, fn string) (outpath string, err error) {
	file, err := ioutil.TempFile("", "shub-")
	if err != nil {
		return
	}

	ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(fn), "."))

	switch ext {
	case "JPG", "JPEG":
		err = jpeg.Encode(file, img, &jpeg.Options{
			Quality: 100, // We don't care about file size, only quality
		})
	case "PNG":
		err = png.Encode(file, img)
	default:
		// Reject the image.
		// Technically, we could work with the image (as it was decoded successfully), but for it to have
		// the correct filename extension we would need to return the correct destination path,
		// Otherwise other parts of the software would become confused.
		err = fmt.Errorf("invalid/unsupported image file extension '%s': use jpg/jpeg or png", ext)
	}
	if err != nil {
		file.Close()
		return
	}

	err = file.Close()
	if err != nil {
		return
	}

	return file.Name(), nil
}

// BetterCover searches a better cover for a song that has artist and album information.
// If no error is returned, it can be found at `path`. From there, it must be copied, not moved!
func (m *Manager) betterCover(artist, album, currentPath string) (path string, err error) {
	if artist == "" || album == "" {
		return
	}

	coverCandidates, ok := m.GetAlbum(artist, album)
	if !ok {
		err = fmt.Errorf("This album doesn't exist yet")
		return
	}

	orig, err := images.Open(currentPath)
	if err != nil {
		return
	}

	origHash, origSize := images.Hash(orig)

	// Try every song in that album
	for _, song := range coverCandidates.Songs {
		// No cover, no luck
		if song.PictureData.Filename == "" {
			continue
		}

		// Open & hash the image we already have
		curr, err := images.Open(song.CoverPath())
		if err != nil {
			return "", err
		}

		// if the current image is not larger than the one that should be replaced, we skip it
		if curr.Bounds().Dx() < orig.Bounds().Dx() {
			continue
		}

		currHash, currSize := images.Hash(curr)

		if images.Similar(origHash, currHash, origSize, currSize) {
			return song.CoverPath(), nil
		}
	}

	err = fmt.Errorf("cannot find any matching images in albums that already exist")
	return
}

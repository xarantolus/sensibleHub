package store

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// getAudioDuration returns the duration of the audio in seconds (also works on videos)
func (m *Manager) getAudioDuration(inputPath string) (durationSeconds float64, err error) {
	if strings.TrimSpace(inputPath) == "" {
		return 0, fmt.Errorf("inputPath is empty")
	}

	cmd := exec.Command(m.cfg.Alternatives.FFprobe, "-i", inputPath, "-show_entries", "format=duration", "-print_format", "json", "-v", "quiet")

	output, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	var out durationInfo
	err = json.NewDecoder(output).Decode(&out)
	if err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil {
		return
	}

	return strconv.ParseFloat(out.Format.Duration, 64)
}

type durationInfo struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

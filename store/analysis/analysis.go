package analysis

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"time"
	"xarantolus/sensibleHub/store/music"
)

// Version identifies the algorithm. Bump it when the descriptors change meaning,
// so stored analyses are recomputed.
const Version = 1

// maxSeconds bounds the work per song; long mixes are analysed from their start.
const maxSeconds = 15 * 60

// Analyze decodes the audio at path between start and end (seconds; end <= 0
// means the whole file) with ffmpeg and computes its descriptors.
func Analyze(ctx context.Context, ffmpeg, path string, start, end float64) (music.Analysis, error) {
	res := music.Analysis{Version: Version, AnalyzedAt: time.Now(), Start: start, End: end}

	feat, err := decodeAndAnalyze(ctx, ffmpeg, path, start, end)
	if err != nil {
		return res, err
	}
	lufs, lra, err := loudness(ctx, ffmpeg, path, start, end)
	if err != nil {
		return res, err
	}

	res.Key, res.Minor, res.KeyStrength = feat.Key, feat.Minor, feat.KeyStrength
	res.Camelot = Camelot(feat.Key, feat.Minor)
	res.BPM, res.BeatStrength = feat.BPM, feat.BeatStrength
	res.LoudnessLUFS, res.LoudnessRange = lufs, lra
	res.EnergyDB, res.EnergyDBStdDev, res.OnsetRate = feat.EnergyDB, feat.EnergyDBStdDev, feat.OnsetRate
	res.Centroid, res.Rolloff, res.Flatness = feat.Centroid, feat.Rolloff, feat.Flatness
	res.Timbre = feat.Timbre
	return res, nil
}

func rangeArgs(start, end float64) []string {
	var args []string
	if start > 0 {
		args = append(args, "-ss", strconv.FormatFloat(start, 'f', 3, 64))
	}
	dur := float64(maxSeconds)
	if end > 0 && end-start < dur {
		dur = end - max(start, 0)
	}
	return append(args, "-t", strconv.FormatFloat(dur, 'f', 3, 64))
}

func decodeAndAnalyze(ctx context.Context, ffmpeg, path string, start, end float64) (Features, error) {
	args := append([]string{"-v", "error", "-nostdin"}, rangeArgs(start, end)...)
	args = append(args, "-i", path, "-vn", "-ac", "1", "-ar", strconv.Itoa(SampleRate), "-f", "f32le", "-")
	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Features{}, err
	}
	if err := cmd.Start(); err != nil {
		return Features{}, err
	}

	a := NewAnalyzer()
	r := bufio.NewReaderSize(stdout, 1<<16)
	raw := make([]byte, 4*4096)
	samples := make([]float32, 4096)
	for {
		n, rerr := io.ReadFull(r, raw)
		n -= n % 4
		for i := 0; i < n/4; i++ {
			samples[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[4*i:]))
		}
		a.Write(samples[:n/4])
		if rerr != nil {
			if !errors.Is(rerr, io.EOF) && !errors.Is(rerr, io.ErrUnexpectedEOF) {
				_ = cmd.Wait()
				return Features{}, rerr
			}
			break
		}
	}
	if err := cmd.Wait(); err != nil {
		return Features{}, fmt.Errorf("decoding audio: %w: %s", err, bytes.TrimSpace(stderr.Bytes()))
	}

	f, ok := a.Features()
	if !ok {
		return Features{}, errors.New("the audio is silent")
	}
	return f, nil
}

var (
	integratedRe = regexp.MustCompile(`I:\s+(-?[\d.]+|-inf)\s+LUFS`)
	rangeRe      = regexp.MustCompile(`LRA:\s+(-?[\d.]+)\s+LU`)
)

// loudness measures integrated loudness (LUFS) and loudness range (LU) per EBU R128.
func loudness(ctx context.Context, ffmpeg, path string, start, end float64) (lufs, lra float64, err error) {
	args := append([]string{"-nostats", "-hide_banner", "-nostdin"}, rangeArgs(start, end)...)
	args = append(args, "-i", path, "-vn", "-filter_complex", "ebur128", "-f", "null", "-")
	out, err := exec.CommandContext(ctx, ffmpeg, args...).CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("measuring loudness: %w", err)
	}
	return parseLoudness(out)
}

func parseLoudness(out []byte) (lufs, lra float64, err error) {
	i := integratedRe.FindAllSubmatch(out, -1)
	r := rangeRe.FindAllSubmatch(out, -1)
	if len(i) == 0 || len(r) == 0 {
		return 0, 0, errors.New("ffmpeg printed no loudness summary")
	}
	if s := string(i[len(i)-1][1]); s == "-inf" {
		lufs = math.Inf(-1)
	} else if lufs, err = strconv.ParseFloat(s, 64); err != nil {
		return 0, 0, err
	}
	if lra, err = strconv.ParseFloat(string(r[len(r)-1][1]), 64); err != nil {
		return 0, 0, err
	}
	if math.IsInf(lufs, -1) {
		return 0, 0, errors.New("the audio is silent")
	}
	return lufs, lra, nil
}

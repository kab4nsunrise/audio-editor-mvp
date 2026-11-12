package audio

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ProcessParams struct {
	Filename  string
	StartTime string
	EndTime   string
	Volume    float64
	FadeIn    float64
	FadeOut   float64
	Format    string
	Normalize bool
}

type Service struct {
	UploadDir string
	mu        sync.Mutex
}

func NewService(uploadDir string) *Service {
	return &Service{UploadDir: uploadDir}
}

func (s *Service) EnsureUploadDir() error {
	return os.MkdirAll(s.UploadDir, 0755)
}

func (s *Service) SaveUpload(r io.Reader, originalName string) (filename string, url string, err error) {
	safe := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitize(originalName))
	dstPath := filepath.Join(s.UploadDir, safe)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, r); err != nil {
		return "", "", err
	}

	return safe, "/uploads/" + safe, nil
}

func (s *Service) Process(p ProcessParams) (outName string, url string, err error) {
	format := strings.ToLower(p.Format)
	switch format {
	case "mp3", "wav", "ogg":
	default:
		format = "mp3"
	}

	if p.Volume <= 0 {
		p.Volume = 1.0
	}
	if p.Volume > 3.0 {
		p.Volume = 3.0
	}

	inputPath := filepath.Join(s.UploadDir, filepath.Base(p.Filename))
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("file not found")
	}

	base := strings.TrimSuffix(filepath.Base(p.Filename), filepath.Ext(p.Filename))
	outName = fmt.Sprintf("processed_%d_%s.%s", time.Now().UnixNano(), base, format)
	outputPath := filepath.Join(s.UploadDir, outName)

	args := buildArgs(inputPath, outputPath, p, format)
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("ffmpeg: %w (%s)", err, string(output))
	}

	return outName, "/uploads/" + outName, nil
}

func buildArgs(input, output string, p ProcessParams, format string) []string {
	args := []string{"-y", "-hide_banner", "-loglevel", "error", "-i", input}

	if p.StartTime != "" {
		args = append(args, "-ss", p.StartTime)
	}
	if p.EndTime != "" {
		args = append(args, "-to", p.EndTime)
	}

	var filters []string
	if p.Normalize {
		filters = append(filters, "loudnorm=I=-16:TP=-1.5:LRA=11")
	}
	if p.Volume != 1.0 {
		filters = append(filters, fmt.Sprintf("volume=%.2f", p.Volume))
	}
	if p.FadeIn > 0 {
		filters = append(filters, fmt.Sprintf("afade=t=in:st=0:d=%.2f", p.FadeIn))
	}
	if p.FadeOut > 0 {
		filters = append(filters, "areverse")
		filters = append(filters, fmt.Sprintf("afade=t=in:st=0:d=%.2f", p.FadeOut))
		filters = append(filters, "areverse")
	}

	if len(filters) > 0 {
		args = append(args, "-af", strings.Join(filters, ","))
	}

	switch format {
	case "wav":
		args = append(args, "-c:a", "pcm_s16le")
	case "ogg":
		args = append(args, "-c:a", "libvorbis", "-q:a", "5")
	default:
		args = append(args, "-c:a", "libmp3lame", "-q:a", "2")
	}

	args = append(args, output)
	return args
}

func (s *Service) CleanupOld(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.UploadDir)
	if err != nil {
		return
	}
	now := time.Now()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > maxAge {
			_ = os.Remove(filepath.Join(s.UploadDir, e.Name()))
		}
	}
}

func (s *Service) StartCleanupLoop(every, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for range ticker.C {
			s.CleanupOld(maxAge)
		}
	}()
}

func sanitize(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	if name == "" || name == "." {
		name = "audio"
	}
	return name
}

func AllowedExt(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".mp3", ".wav", ".m4a", ".aac", ".flac", ".ogg", ".opus", ".wma", ".webm":
		return true
	}
	return false
}

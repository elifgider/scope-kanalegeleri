package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type UploadService struct {
	uploadsDir string
}

func NewUploadService(uploadsDir string) *UploadService {
	return &UploadService{uploadsDir: uploadsDir}
}

func (s *UploadService) EnsureUploadsDir() error {
	return os.MkdirAll(s.uploadsDir, 0o755)
}

func (s *UploadService) SaveImage(r *http.Request) (string, error) {
	const maxUploadSize = 10 << 20

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return "", err
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(data)
	allowedTypes := []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	if !slices.Contains(allowedTypes, contentType) {
		return "", ErrUnsupportedImageType
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = extensionForContentType(contentType)
	}
	if ext == "" {
		return "", ErrInvalidFileExtension
	}

	if err := s.EnsureUploadsDir(); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), slugifyFilename(strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))), ext)
	path := filepath.Join(s.uploadsDir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}

func extensionForContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}

func slugifyFilename(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-", "(", "", ")", "", "[", "", "]", "")
	name = replacer.Replace(name)

	var buffer bytes.Buffer
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			buffer.WriteRune(char)
		}
	}
	if buffer.Len() == 0 {
		return "gorsel"
	}
	return buffer.String()
}

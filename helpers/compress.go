package helpers

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func CompressFile(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return compressJPEG(path, 75)
	case ".png":
		return compressPNG(path)
	case ".pdf":
		return compressPDF(path)
	}
	return nil
}

func compressJPEG(path string, quality int) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("gagal buka file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("gagal decode gambar: %w", err)
	}
	f.Close()

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("gagal buka file output: %w", err)
	}
	defer out.Close()

	return jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
}

func compressPNG(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("gagal buka file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("gagal decode gambar: %w", err)
	}
	f.Close()

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("gagal buka file output: %w", err)
	}
	defer out.Close()

	enc := &png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(out, img)
}

func compressPDF(path string) error {
	err := api.OptimizeFile(path, path, nil)
	if err != nil {
		return fmt.Errorf("gagal kompres PDF: %w", err)
	}
	return nil
}

package services

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type QuestionAssetImportResult struct {
	UploadedCount int      `json:"uploadedCount"`
	SkippedCount  int      `json:"skippedCount"`
	Files         []string `json:"files"`
	Errors        []string `json:"errors"`
}

var allowedQuestionAssetExtensions = map[string]bool{
	".gif":  true,
	".jpeg": true,
	".jpg":  true,
	".png":  true,
	".svg":  true,
	".webp": true,
}

func ImportQuestionAssetZip(filename string, raw []byte, assetDir string) (QuestionAssetImportResult, error) {
	result := QuestionAssetImportResult{
		Files:  []string{},
		Errors: []string{},
	}

	if strings.ToLower(filepath.Ext(filename)) != ".zip" {
		return result, errors.New("file asset gambar harus berformat .zip")
	}
	if strings.TrimSpace(assetDir) == "" {
		return result, errors.New("folder asset gambar belum dikonfigurasi")
	}

	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return result, fmt.Errorf("file ZIP tidak valid: %w", err)
	}

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		assetPath, err := NormalizeQuestionAssetPath(file.Name)
		if err != nil {
			result.SkippedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", file.Name, err.Error()))
			continue
		}

		if !allowedQuestionAssetExtensions[strings.ToLower(filepath.Ext(assetPath))] {
			result.SkippedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: format gambar tidak didukung", file.Name))
			continue
		}

		source, err := file.Open()
		if err != nil {
			result.SkippedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: file tidak bisa dibuka", file.Name))
			continue
		}

		targetPath := filepath.Join(assetDir, filepath.FromSlash(assetPath))
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			source.Close()
			return result, err
		}

		target, err := os.Create(targetPath)
		if err != nil {
			source.Close()
			return result, err
		}

		_, copyErr := io.Copy(target, source)
		closeErr := target.Close()
		source.Close()
		if copyErr != nil {
			return result, copyErr
		}
		if closeErr != nil {
			return result, closeErr
		}

		result.UploadedCount++
		result.Files = append(result.Files, "/api/v1/question-assets/"+assetPath)
	}

	if result.UploadedCount == 0 {
		return result, errors.New("tidak ada gambar valid yang berhasil diupload")
	}

	return result, nil
}

func NormalizeQuestionAssetPath(value string) (string, error) {
	normalized := filepath.ToSlash(strings.TrimSpace(value))
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimPrefix(normalized, "question-assets/")
	normalized = strings.TrimPrefix(normalized, "api/v1/question-assets/")
	normalized = filepath.ToSlash(filepath.Clean(normalized))
	normalized = strings.TrimPrefix(normalized, "./")

	if normalized == "." || normalized == "" {
		return "", errors.New("path gambar kosong")
	}
	if strings.HasPrefix(normalized, "../") || strings.Contains(normalized, "/../") || filepath.IsAbs(normalized) {
		return "", errors.New("path gambar tidak aman")
	}
	if strings.Contains(normalized, "\\") {
		return "", errors.New("path gambar tidak valid")
	}

	return normalized, nil
}

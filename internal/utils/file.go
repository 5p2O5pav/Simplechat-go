package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func GenerateUniqueFileName(ext string) string {
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	return fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), hex.EncodeToString(randBytes), ext)
}

func SaveUploadedFile(data io.Reader, destPath string) (int64, error) {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, err
	}
	f, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := io.Copy(f, data)
	return n, err
}

func DeleteFile(path string) error {
	return os.Remove(path)
}

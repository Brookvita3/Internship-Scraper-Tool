package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Thay thế ký tự không hợp lệ
func SanitizeFolderName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}

// Download file từ URL
func DownloadFile(url string, dest string) error {
	fmt.Println("Downloading", url, "to", dest)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

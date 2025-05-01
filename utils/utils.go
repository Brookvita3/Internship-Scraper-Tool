package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"scarpe-intern/models"
	"strings"
)

// Thay thế ký tự không hợp lệ
func sanitizeFolderName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}

// Download file từ URL
func downloadFile(url string, dest string) error {
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

func FetchCompanyDetails(url string) (models.CompanyDetails, error) {
	res, err := http.Get(url)
	if err != nil {
		return models.CompanyDetails{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return models.CompanyDetails{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	var response map[string]any
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return models.CompanyDetails{}, err
	}

	item := response["item"].(map[string]any)
	company := models.CompanyDetails{
		Fullname:               item["fullname"].(string),
		MaxAcceptedStudent:     int(item["maxAcceptedStudent"].(float64)),
		StudentRegister:        int(item["studentRegister"].(float64)),
		StudentAccepted:        int(item["studentAccepted"].(float64)),
		SubscribeAcceptedEmail: item["subscribeAcceptedEmail"].(bool),
		MaxRegister:            int(item["maxRegister"].(float64)),
	}

	if filesRaw, ok := item["internshipFiles"].([]any); ok {
		for _, f := range filesRaw {
			fileMap := f.(map[string]any)
			company.InternshipFiles = append(company.InternshipFiles, models.InternshipFile{
				Name: fileMap["name"].(string),
				Path: fileMap["path"].(string),
			})
		}
	}

	return company, nil
}

func IsAvailable(company models.CompanyDetails) bool {
	if company.SubscribeAcceptedEmail &&
		company.StudentRegister < company.MaxRegister &&
		company.StudentAccepted < company.MaxAcceptedStudent {
		return true
	}
	return false
}

func SaveInfoCompany(company models.CompanyDetails, errChan chan<- error) {
	// Tạo thư mục theo tên công ty
	folderName := sanitizeFolderName(company.Fullname)
	dirPath := "company/" + folderName
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		errChan <- fmt.Errorf("failed to create folder for %s: %v", company.Fullname, err)
		return
	}

	// Tải các file
	for _, file := range company.InternshipFiles {
		fileURL := os.Getenv("BASE_URL") + file.Path
		log.Printf("Downloading file from URL: %s", fileURL)
		destPath := dirPath + "/" + file.Name
		if err := downloadFile(fileURL, destPath); err != nil {
			errChan <- fmt.Errorf("failed to download file %s: %v", file.Name, err)
		}
	}
}

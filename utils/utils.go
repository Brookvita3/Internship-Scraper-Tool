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
	"sync"
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

func fetchCompanyDetails(url string) (models.CompanyDetails, error) {
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

func isAvailable(company models.CompanyDetails) bool {
	if company.StudentRegister < company.MaxRegister &&
		company.StudentAccepted < company.MaxAcceptedStudent {
		return true
	}
	return false
}

func saveInfoCompany(company models.CompanyDetails, errChan chan<- error) {
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

// Hàm lấy danh sách ID từ URL đầu tiên
func GetCompanyIDs() ([]string, error) {
	listCompanyURL := os.Getenv("LIST_COMPANY_URL")
	res, err := http.Get(listCompanyURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	var response map[string]any
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	var ids []string

	items := response["items"].([]any)
	for _, item := range items {
		id := item.(map[string]any)["_id"].(string)
		ids = append(ids, id)
	}

	return ids, nil
}

func GetCompanyDetailsWorker(ids <-chan string, errChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	companyDetailsURL := os.Getenv("COMPANY_DETAILS_URL")
	for id := range ids {

		// Fetch thông tin chi tiết công ty
		company, err := fetchCompanyDetails(companyDetailsURL + "/" + id)
		if err != nil {
			errChan <- fmt.Errorf("failed to fetch company details for %s: %v", id, err)
			continue
		}

		// Tạo thư mục cho công ty thõa mãn điều kiện
		if isAvailable(company) {
			saveInfoCompany(company, errChan)
		}
	}
}

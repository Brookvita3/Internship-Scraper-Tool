package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"scarpe-intern/models"
	"scarpe-intern/utils"
	"sync"
)

// Đường dẫn API
const (
	listCompanyURL    = "https://internship.cse.hcmut.edu.vn/home/company/all?condition="
	companyDetailsURL = "https://internship.cse.hcmut.edu.vn/home/company/id"
	baseURL           = "https://internship.cse.hcmut.edu.vn"
)

// Hàm lấy danh sách ID từ URL đầu tiên
func getCompanyIDs() ([]string, error) {
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

func getCompanyDetailsWorker(ids <-chan string, errChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for id := range ids {
		url := companyDetailsURL + "/" + id
		res, err := http.Get(url)
		if err != nil {
			errChan <- err
			continue
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			errChan <- fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
			continue
		}

		var response map[string]any
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			errChan <- err
			continue
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

		// Chỉ lấy những công ty có điều kiện
		if company.SubscribeAcceptedEmail &&
			company.StudentRegister < company.MaxRegister &&
			company.StudentAccepted < company.MaxAcceptedStudent {

			// Tạo thư mục theo tên công ty
			folderName := utils.SanitizeFolderName(company.Fullname)
			dirPath := filepath.Join("company", folderName)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				errChan <- fmt.Errorf("failed to create folder for %s: %v", company.Fullname, err)
				continue
			}

			// Tải các file
			for _, file := range company.InternshipFiles {
				fileURL := baseURL + file.Path
				destPath := filepath.Join(dirPath, file.Name)
				if err := utils.DownloadFile(fileURL, destPath); err != nil {
					errChan <- fmt.Errorf("failed to download file %s: %v", fileURL, err)
				}
			}
		}
	}
}

func main() {

	// Lấy danh sách ID
	ids, err := getCompanyIDs()
	if err != nil {
		log.Fatalf("Error fetching company IDs: %v", err)
	}

	var wg sync.WaitGroup
	idsChan := make(chan string)
	detailsChan := make(chan models.CompanyDetails)
	errorChan := make(chan error)

	// Tạo worker
	for range 10 {
		wg.Add(1)
		go getCompanyDetailsWorker(idsChan, errorChan, &wg)
	}

	// Gửi ID vào channel
	go func() {
		for _, id := range ids[:] { // ví dụ chỉ lấy 10 ID đầu tiên
			idsChan <- id
		}
		close(idsChan)
	}()

	// Đóng result/error channel sau khi tất cả goroutine xong
	go func() {
		wg.Wait()
		close(detailsChan)
		close(errorChan)
	}()

	// In lỗi nếu có
	for err := range errorChan {
		log.Println("Lỗi:", err)
	}

	fmt.Println("Dữ liệu đã được cập nhật thành công.")
}

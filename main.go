package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"scarpe-intern/models"
	"scarpe-intern/utils"
	"sync"
)

// Đường dẫn API
var (
	listCompanyURL    string
	companyDetailsURL string
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

		// Fetch thông tin chi tiết công ty
		company, err := utils.FetchCompanyDetails(companyDetailsURL + "/" + id)
		if err != nil {
			errChan <- fmt.Errorf("failed to fetch company details for %s: %v", id, err)
			continue
		}

		// Tạo thư mục cho công ty thõa mãn điều kiện
		if utils.IsAvailable(company) {
			utils.SaveInfoCompany(company, errChan)
		}
	}
}

func main() {

	err := os.RemoveAll("company")
	if err != nil {
		log.Fatalf("Error removing directory: %v", err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	listCompanyURL = os.Getenv("LIST_COMPANY_URL")
	companyDetailsURL = os.Getenv("COMPANY_DETAILS_URL")

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
		for _, id := range ids[:] {
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
		log.Println("Error:", err)
	}

	fmt.Println("Dữ liệu đã được cập nhật thành công.")
}

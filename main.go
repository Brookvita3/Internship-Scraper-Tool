package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"scarpe-intern/utils"
	"sync"
)

// Đường dẫn API
var (
	listCompanyURL    string
	companyDetailsURL string
)

func runWorkers(ids []string, errors *[]error, mu *sync.Mutex) {
	var wg sync.WaitGroup
	idsChan := make(chan string)
	errorChan := make(chan error)

	// Tạo worker
	for range 10 {
		wg.Add(1)
		go utils.GetCompanyDetailsWorker(idsChan, errorChan, &wg)
	}

	// Gửi ID vào channel
	go func() {
		for _, id := range ids {
			idsChan <- id
		}
		close(idsChan)
	}()

	// Thu thập lỗi
	go func() {
		for err := range errorChan {
			mu.Lock()
			*errors = append(*errors, err)
			mu.Unlock()
		}
	}()

	// Chờ tất cả worker hoàn thành và đóng errorChan
	wg.Wait()
	close(errorChan)

}

func main() {

	// Xóa thư mục company cũ
	err := os.RemoveAll("company")
	if err != nil {
		log.Fatalf("Error removing directory: %v", err)
	}

	// Tải biến môi trường từ file .env
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	listCompanyURL = os.Getenv("LIST_COMPANY_URL")
	companyDetailsURL = os.Getenv("COMPANY_DETAILS_URL")

	// Lấy danh sách ID
	ids, err := utils.GetCompanyIDs()
	if err != nil {
		log.Fatalf("Error fetching company IDs: %v", err)
	}

	// Thu thập lỗi
	var errors []error
	var mu sync.Mutex // Mutex để bảo vệ slice errors khi append đồng thời

	// Chạy workers và xử lý
	runWorkers(ids, &errors, &mu)

	// Báo cáo kết quả
	if len(errors) > 0 {
		log.Println("Có lỗi xảy ra trong quá trình xử lý:")
		for i, err := range errors {
			log.Printf("Lỗi %d: %v", i+1, err)
		}
		log.Fatal("Chương trình kết thúc với lỗi.")
	}

	fmt.Println("Dữ liệu đã được cập nhật thành công.")
}

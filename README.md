# Internship Scraper Tool

Một công cụ nhẹ viết bằng Golang dùng để thu thập dữ liệu các công ty tuyển thực tập từ Cổng Thực Tập của Khoa Khoa học & Kỹ thuật Máy tính - Đại học Bách Khoa TP.HCM (HCMUT CSE) và tải xuống các tài liệu liên quan (5/2025).

## 📌 Feature

- Lấy danh sách tất cả ID công ty từ hệ thống thực tập
- Lọc công ty dựa trên các điều kiện:
  - Vẫn cho phép đăng ký
  - Chưa vượt quá số lượng sinh viên được chấp nhận
  - Có bật gửi email xác nhận (`subscribeAcceptedEmail`)
- Tải xuống các file thực tập đính kèm
- Sắp xếp file vào các thư mục theo tên công ty

## ⚙️ Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/internship-scraper.git
cd internship-scraper
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Run the Scraper

```bash
go run main.go
```

Tất cả công ty đủ điều kiện và file đính kèm sẽ được tải về thư mục company/.

package models

// Cấu trúc dữ liệu cho ID từ response đầu tiên
type CompanyItem struct {
	ID string `json:"_id"`
}

type InternshipFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Cấu trúc dữ liệu cho thông tin chi tiết công ty
type CompanyDetails struct {
	Fullname               string           `json:"fullname"`
	MaxAcceptedStudent     int              `json:"maxAcceptedStudent"`
	StudentRegister        int              `json:"studentRegister"`
	StudentAccepted        int              `json:"studentAccepted"`
	SubscribeAcceptedEmail bool             `json:"subscribeAcceptedEmail"`
	MaxRegister            int              `json:"maxRegister"`
	InternshipFiles        []InternshipFile `json:"internshipFiles"`
}

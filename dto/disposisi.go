package dto

import "time"

// ===== REQUEST DTOs =====

// CreateForwardRequest - request untuk forward disposisi
type CreateForwardRequest struct {
	ToUserID    uint   `json:"to_user_id" binding:"required"`
	Catatan     string `json:"catatan"`
	Sifat       string `json:"sifat"` // penting, biasa, rahasia
	ActionType  string `json:"action_type" binding:"required"` // forward, reject, complete
}

// CompleteDisposisiRequest - request untuk mark disposisi as complete
type CompleteDisposisiRequest struct {
	TanggapanSaran       string `json:"tanggapan_saran"`
	ProsesLanjut         string `json:"proses_lanjut"`
	KoordinasiKonfirmasi string `json:"koordinasi_konfirmasi"`
	Catatan              string `json:"catatan"`
}

// UpdateDisposisiRequest - request untuk update disposisi
type UpdateDisposisiRequest struct {
	Catatan string `json:"catatan"`
	Sifat   string `json:"sifat"`
}

// ===== RESPONSE DTOs =====

// UserBasicInfo - info user simplified
type UserBasicInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

// JabatanBasicInfo - info jabatan simplified
type JabatanBasicInfo struct {
	ID          uint   `json:"id"`
	NamaJabatan string `json:"nama_jabatan"`
}

// DisposisiResponse - response detail disposisi
type DisposisiResponse struct {
	ID                uint              `json:"id"`
	SuratMasukID      uint              `json:"surat_masuk_id"`
	FromUser          UserBasicInfo     `json:"from_user"`
	ToUser            UserBasicInfo     `json:"to_user"`
	ParentDisposisiID *uint             `json:"parent_disposisi_id"`
	Level             int               `json:"level"`
	Status            string            `json:"status"`
	Catatan           string            `json:"catatan"`
	Dibaca            bool              `json:"dibaca"`
	Sifat             string            `json:"sifat"`
	TanggapanSaran    string            `json:"tanggapan_saran"`
	ProsesLanjut      string            `json:"proses_lanjut"`
	KoordinasiKonfirmasi string          `json:"koordinasi_konfirmasi"`
	BacaAt            *time.Time        `json:"baca_at"`
	CompleteAt        *time.Time        `json:"complete_at"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	ChildCount        int               `json:"child_count"` // jumlah child disposisi
}

// InboxItemResponse - response untuk inbox list
type InboxItemResponse struct {
	ID               uint          `json:"id"`
	SuratMasukID     uint          `json:"surat_masuk_id"`
	FromUser         UserBasicInfo `json:"from_user"`
	Status           string        `json:"status"`
	Catatan          string        `json:"catatan"`
	Dibaca           bool          `json:"dibaca"`
	Sifat            string        `json:"sifat"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	SuratNomor       string        `json:"surat_nomor"`   // dari surat_masuk
	SuratPerihal     string        `json:"surat_perihal"` // dari surat_masuk
}

// SentItemResponse - response untuk sent items list
type SentItemResponse struct {
	ID              uint          `json:"id"`
	SuratMasukID    uint          `json:"surat_masuk_id"`
	ToUser          UserBasicInfo `json:"to_user"`
	Status          string        `json:"status"`
	Catatan         string        `json:"catatan"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	ChildCount      int           `json:"child_count"`       // jumlah forward lanjut
	CompletedCount  int           `json:"completed_count"`   // berapa yang sudah completed
}

// DisposisiTreeNode - node untuk tree structure
type DisposisiTreeNode struct {
	ID               uint                  `json:"id"`
	FromUser         UserBasicInfo         `json:"from_user"`
	ToUser           UserBasicInfo         `json:"to_user"`
	Level            int                   `json:"level"`
	Status           string                `json:"status"`
	Catatan          string                `json:"catatan"`
	CreatedAt        time.Time             `json:"created_at"`
	CompleteAt       *time.Time            `json:"complete_at"`
	Children         []DisposisiTreeNode   `json:"children"`
}

// HistoryResponse - response untuk full history surat
type HistoryResponse struct {
	SuratMasukID uint                 `json:"surat_masuk_id"`
	SuratNomor   string               `json:"surat_nomor"`
	SuratPerihal string               `json:"surat_perihal"`
	TanggalSurat *time.Time           `json:"tanggal_surat"`
	AsalSurat    string               `json:"asal_surat"`
	RootDisposisi DisposisiTreeNode   `json:"root_disposisi"`
	TotalForward int                  `json:"total_forward"`
	Status       string               `json:"status"` // based on root disposisi status
}

// PaginationMeta - metadata untuk pagination
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// InboxListResponse - response untuk inbox dengan pagination
type InboxListResponse struct {
	Data       []InboxItemResponse `json:"data"`
	Pagination PaginationMeta      `json:"pagination"`
}

// SentListResponse - response untuk sent items dengan pagination
type SentListResponse struct {
	Data       []SentItemResponse `json:"data"`
	Pagination PaginationMeta     `json:"pagination"`
}

// ForwardValidationResponse - response untuk validasi forward
type ForwardValidationResponse struct {
	CanForward      bool     `json:"can_forward"`
	Reason          string   `json:"reason"`
	AllowedJabatans []string `json:"allowed_jabatans"`
}

// ===== ERROR RESPONSE DTOs =====

// ErrorResponse - response untuk error
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ValidationErrorDetail - detail untuk validation error
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
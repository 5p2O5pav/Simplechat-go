package models

type Message struct {
	ID       string `json:"id"`
	Sid      string `json:"-"`
	Role     string `json:"role"`
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
	FileURL  string `json:"fileUrl,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Time     int64  `json:"time"`
	Metadata string `json:"metadata,omitempty"` // JSON string for extra fields
}

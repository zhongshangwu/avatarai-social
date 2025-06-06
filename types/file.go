package types

type UploadFile struct {
	ID        string `json:"id"`
	Size      int64  `json:"size"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`
	CID       string `json:"cid"`
	URL       string `json:"url"`
	CreatedBy string `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
}

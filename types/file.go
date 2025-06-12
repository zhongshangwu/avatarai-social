package types

type UploadFile struct {
	ID        string `json:"id"`
	Size      int64  `json:"size"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	MimeType  string `json:"mimeType"`
	CID       string `json:"cid"`
	URL       string `json:"url"`
	BlobCID   string `json:"blobCid"`
	CreatedBy string `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
}

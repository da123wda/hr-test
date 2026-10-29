package dto

type UploadFileResp struct {
	FileId       string `json:"fileId"`
	OriginalName string `json:"originalName"`
	FileSize     int64  `json:"fileSize"`
	MimeType     string `json:"mimeType"`
	DownloadUrl  string `json:"downloadUrl"`
}

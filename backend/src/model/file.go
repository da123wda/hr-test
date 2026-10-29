package model

import "time"

// FileMetadata 上传文件的受控元数据 (Q3)
type FileMetadata struct {
	FileId       string    `json:"fileId" xorm:"varchar(64) pk 'file_id'"`
	OriginalName string    `json:"originalName" xorm:"varchar(255) notnull 'original_name'"`
	StoredName   string    `json:"storedName" xorm:"varchar(255) notnull 'stored_name'"`
	FileSize     int64     `json:"fileSize" xorm:"bigint notnull 'file_size'"`
	MimeType     string    `json:"mimeType" xorm:"varchar(128) notnull 'mime_type'"`
	Sha256       string    `json:"sha256" xorm:"varchar(64) notnull 'sha256'"`
	DownloadUrl  string    `json:"downloadUrl" xorm:"varchar(512) notnull 'download_url'"`
	FormId       string    `json:"formId,omitempty" xorm:"varchar(64) 'form_id'"`
	UploadedBy   string    `json:"uploadedBy,omitempty" xorm:"varchar(64) 'uploaded_by'"`
	CreatedAt    time.Time `json:"createdAt" xorm:"created 'created_at'"`
}

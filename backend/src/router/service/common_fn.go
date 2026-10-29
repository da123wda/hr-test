package service

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
)

func calcFileSha256(file multipart.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	_, _ = file.Seek(0, io.SeekStart)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

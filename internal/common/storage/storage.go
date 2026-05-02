package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Storage defines the interface for file storage backends.
type Storage interface {
	// Upload stores a file and returns its public URL.
	Upload(objectKey string, reader io.Reader, contentType string) (string, error)
	// Delete removes a file.
	Delete(objectKey string) error
	// GetURL returns the public URL for an object key.
	GetURL(objectKey string) string
}

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	basePath string // e.g. "/data/uploads"
	baseURL  string // e.g. "http://localhost:8001/uploads"
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath, baseURL: baseURL}
}

func (s *LocalStorage) Upload(objectKey string, reader io.Reader, contentType string) (string, error) {
	fullPath := filepath.Join(s.basePath, objectKey)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create dir failed: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("write file failed: %w", err)
	}

	return s.GetURL(objectKey), nil
}

func (s *LocalStorage) Delete(objectKey string) error {
	fullPath := filepath.Join(s.basePath, objectKey)
	return os.Remove(fullPath)
}

func (s *LocalStorage) GetURL(objectKey string) string {
	return strings.TrimRight(s.baseURL, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

// MinIOStorage stores files in MinIO (S3-compatible).
type MinIOStorage struct {
	endpoint  string // e.g. "minio:9000"
	bucket    string // e.g. "aqicloud"
	accessKey string
	secretKey string
	useSSL    bool
	publicURL string // e.g. "http://minio:9000/aqicloud"
}

func NewMinIOStorage(endpoint, bucket, accessKey, secretKey string, useSSL bool, publicURL string) *MinIOStorage {
	return &MinIOStorage{
		endpoint:  endpoint,
		bucket:    bucket,
		accessKey: accessKey,
		secretKey: secretKey,
		useSSL:    useSSL,
		publicURL: publicURL,
	}
}

func (s *MinIOStorage) Upload(objectKey string, reader io.Reader, contentType string) (string, error) {
	// Use S3 PutObject via HTTP (no external SDK dependency)
	// This implements AWS S3 Signature V4 for MinIO compatibility
	url := fmt.Sprintf("http://%s/%s/%s", s.endpoint, s.bucket, objectKey)
	if s.useSSL {
		url = fmt.Sprintf("https://%s/%s/%s", s.endpoint, s.bucket, objectKey)
	}

	// Read all data for content-length
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read data failed: %w", err)
	}

	req, err := newS3Request("PUT", url, data, s.bucket, objectKey, s.accessKey, s.secretKey, contentType)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	resp, err := httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	return s.GetURL(objectKey), nil
}

func (s *MinIOStorage) Delete(objectKey string) error {
	url := fmt.Sprintf("http://%s/%s/%s", s.endpoint, s.bucket, objectKey)
	if s.useSSL {
		url = fmt.Sprintf("https://%s/%s/%s", s.endpoint, s.bucket, objectKey)
	}

	req, err := newS3Request("DELETE", url, nil, s.bucket, objectKey, s.accessKey, s.secretKey, "")
	if err != nil {
		return err
	}

	resp, err := httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return fmt.Errorf("delete failed: status=%d", resp.StatusCode)
	}
	return nil
}

func (s *MinIOStorage) GetURL(objectKey string) string {
	return strings.TrimRight(s.publicURL, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

// GenerateObjectKey creates a unique object key based on date and hash.
func GenerateObjectKey(originalFilename string, hash string) string {
	ext := filepath.Ext(originalFilename)
	date := time.Now().Format("2006/01/02")
	return fmt.Sprintf("user/%s/%s%s", date, hash, ext)
}

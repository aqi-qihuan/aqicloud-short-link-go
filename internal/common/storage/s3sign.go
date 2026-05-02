package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// newS3Request creates an HTTP request with AWS S3 Signature V4.
// Minimal implementation for MinIO compatibility.
func newS3Request(method, url string, body []byte, bucket, objectKey, accessKey, secretKey, contentType string) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	dateStr := now.Format("20060102T150405Z")
	dateShort := now.Format("20060102")

	req.Header.Set("Host", req.Host)
	req.Header.Set("X-Amz-Date", dateStr)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Compute payload hash
	payloadHash := sha256Hex(body)

	// Canonical request
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-date:%s\n", req.Host, dateStr)
	signedHeaders := "host;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n/%s/%s\n\n%s\n%s\n%s",
		method, bucket, objectKey,
		canonicalHeaders, signedHeaders, payloadHash)

	// String to sign
	credentialScope := fmt.Sprintf("%s/us-east-1/s3/aws4_request", dateShort)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		dateStr, credentialScope, sha256Hex([]byte(canonicalRequest)))

	// Signing key
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), dateShort)
	regionKey := hmacSHA256(dateKey, "us-east-1")
	serviceKey := hmacSHA256(regionKey, "s3")
	signingKey := hmacSHA256(serviceKey, "aws4_request")

	signature := fmt.Sprintf("%x", hmacSHA256(signingKey, stringToSign))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)

	return req, nil
}

func sha256Hex(data []byte) string {
	if data == nil {
		data = []byte{}
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

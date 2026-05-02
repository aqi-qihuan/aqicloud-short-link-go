package component

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
)

// ensurePEM wraps a raw base64 key with PEM headers if needed.
func ensurePEM(key, header string) string {
	if strings.Contains(key, "-----BEGIN") {
		return key
	}
	key = strings.TrimSpace(key)
	return "-----BEGIN " + header + "-----\n" + key + "\n-----END " + header + "-----"
}

// RSASign signs data with a PKCS8 private key using the specified hash.
func RSASign(data string, privateKeyPEM string, hash crypto.Hash) (string, error) {
	privateKeyPEM = ensurePEM(privateKeyPEM, "PRIVATE KEY")
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse private key PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	privKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	h := hash.New()
	h.Write([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, hash, h.Sum(nil))
	if err != nil {
		return "", fmt.Errorf("RSA sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// RSAVerify verifies an RSA signature with a public key.
func RSAVerify(data, signatureB64, publicKeyPEM string, hash crypto.Hash) (bool, error) {
	publicKeyPEM = ensurePEM(publicKeyPEM, "PUBLIC KEY")
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false, fmt.Errorf("failed to parse public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("parse public key: %w", err)
	}
	pubKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("not an RSA public key")
	}

	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}

	h := hash.New()
	h.Write([]byte(data))
	err = rsa.VerifyPKCS1v15(pubKey, hash, h.Sum(nil), sig)
	return err == nil, nil
}

// AES256GCMDecrypt decrypts an AES-256-GCM encrypted ciphertext.
// Used for WeChat Pay V3 notification resource decryption.
func AES256GCMDecrypt(ciphertextB64, key, nonce, associatedData string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("new AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		return nil, fmt.Errorf("new GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, []byte(nonce), ciphertext, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("GCM decrypt: %w", err)
	}
	return plaintext, nil
}

// --- Alipay helpers ---

// buildAlipaySignContent builds the string to sign from sorted key=value pairs.
func buildAlipaySignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "&")
}

// signAlipayRSA2 signs Alipay request params with RSA2 (SHA256WithRSA).
func signAlipayRSA2(params map[string]string, privateKeyPEM string) (string, error) {
	// sign and sign_type are excluded from the sign content
	delete(params, "sign")
	delete(params, "sign_type")
	content := buildAlipaySignContent(params)
	return RSASign(content, privateKeyPEM, crypto.SHA256)
}

// verifyAlipayRSA2 verifies an Alipay callback signature.
func verifyAlipayRSA2(params map[string]string, sign, publicKeyPEM string) (bool, error) {
	// Remove sign and sign_type from verification content
	delete(params, "sign")
	delete(params, "sign_type")
	content := buildAlipaySignContent(params)
	return RSAVerify(content, sign, publicKeyPEM, crypto.SHA256)
}

// --- WeChat Pay V2 helpers ---

// signWechatV2 generates an MD5 signature for WeChat Pay V2 API.
func signWechatV2(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	parts = append(parts, "key="+apiKey)
	signStr := strings.Join(parts, "&")
	h := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

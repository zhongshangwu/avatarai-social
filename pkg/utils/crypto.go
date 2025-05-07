package utils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/multiformats/go-multibase"
)

func GenerateDIDKey() (string, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	multicodecPrefix := []byte{0xed, 0x01} // Ed25519 公钥的多编码前缀
	multicodecKey := append(multicodecPrefix, publicKey...)

	encoded, err := multibase.Encode(multibase.Base58BTC, multicodecKey)
	if err != nil {
		return "", nil, fmt.Errorf("编码公钥失败: %w", err)
	}

	didKey := fmt.Sprintf("did:key:z%s", encoded)

	return didKey, privateKey, nil
}

func GenerateCode() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	code := base64.RawURLEncoding.EncodeToString(b)
	return code, nil
}

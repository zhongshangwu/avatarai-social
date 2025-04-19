package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

func main() {
	didKey, prvKey, err := utils.GenerateDIDKey()
	if err != nil {
		panic(err)
	}
	fmt.Printf("didKey: %s\n", didKey)
	fmt.Printf("prvKey: %s\n", prvKey)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// 创建 JWK
	key := jose.JSONWebKey{
		Key:       privateKey,
		KeyID:     fmt.Sprintf("demo-%d", time.Now().Unix()),
		Algorithm: "ES256",
		Use:       "sig",
	}

	// 转换为 JSON
	jsonKey, err := key.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", jsonKey)

	var jwk jose.JSONWebKey
	err = jwk.UnmarshalJSON(jsonKey)
	if err != nil {
		panic(err)
	}

	pubKey, err := jwk.Public().MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", pubKey)

}

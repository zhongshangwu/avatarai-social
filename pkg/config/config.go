package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/labstack/gommon/log"
)

type SocialConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	APP      APPConfig      `mapstructure:"app"`
	ATP      ATPConfig      `mapstructure:"atp"`
	Avatar   AvatarConfig   `mapstructure:"avatar"`
	Security SecurityConfig `mapstructure:"security"` // 新增 security
}

type SecurityConfig struct {
	RSAPrivateKey string `mapstructure:"rsa_private_key"` // RSA 私钥，PEM 格式
}

type ServerConfig struct {
	HTTP     HTTPConfig    `mapstructure:"http"`
	Metrics  MetricsConfig `mapstructure:"metrics"`
	AdminKey string        `mapstructure:"admin_key"`
	Domain   string        `mapstructure:"domain"`
}

type HTTPConfig struct {
	Address      string        `mapstructure:"address"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type MetricsConfig struct {
	Address string `mapstructure:"address"`
}

type DatabaseConfig struct {
	Driver             string        `mapstructure:"driver"`
	DSN                string        `mapstructure:"dsn"`
	MaxConnections     int           `mapstructure:"max_connections"`
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	ConnectionLifetime time.Duration `mapstructure:"connection_lifetime"`
	ConnectionTimeout  time.Duration `mapstructure:"connection_timeout"`
}

type StorageConfig struct {
	DataDir string `mapstructure:"data_dir"`
}

type APPConfig struct {
	BundleID string `mapstructure:"bundle_id"`
}

type ATPConfig struct {
	Service         string `mapstructure:"service"`
	ClientJWKSecret string `mapstructure:"client_jwk_secret"`
}

type AvatarConfig struct {
	LLM   LLMConfig    `mapstructure:"llm"`
	Tools []ToolConfig `mapstructure:"tools"`
}

type LLMConfig struct {
	APIURL   string `mapstructure:"api_url"`
	Model    string `mapstructure:"model"`
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
}

type ToolConfig struct {
	ID string `mapstructure:"id"`
}

func (atp *ATPConfig) ClientPubJWKMap() map[string]interface{} {
	var jwk jose.JSONWebKey
	err := jwk.UnmarshalJSON([]byte(atp.ClientJWKSecret))
	if err != nil {
		return nil
	}

	pubKey, err := jwk.Public().MarshalJSON()
	if err != nil {
		return nil
	}

	var pubJWK map[string]interface{}
	err = json.Unmarshal(pubKey, &pubJWK)
	if err != nil {
		return nil
	}

	return pubJWK
}

func (atp *ATPConfig) ClientSecretJWK() jose.JSONWebKey {
	var jwk jose.JSONWebKey
	err := jwk.UnmarshalJSON([]byte(atp.ClientJWKSecret))
	if err != nil {
		return jose.JSONWebKey{}
	}
	return jwk
}

func (atp *ATPConfig) ClientSecretKey() interface{} {
	var jwk jose.JSONWebKey
	err := jwk.UnmarshalJSON([]byte(atp.ClientJWKSecret))
	if err != nil {
		return nil
	}
	return jwk.Key
}

func (atp *ATPConfig) ClientSecretPubKey() interface{} {
	var jwk jose.JSONWebKey
	err := jwk.UnmarshalJSON([]byte(atp.ClientJWKSecret))
	if err != nil {
		return nil
	}

	return jwk.Public().Key
}

func (s *SecurityConfig) GetRSAPrivateKey() interface{} {
	log.Infof("GetRSAPrivateKey， RSAPrivateKey: %s", s.RSAPrivateKey)
	block, _ := pem.Decode([]byte(s.RSAPrivateKey))
	if block == nil {
		log.Errorf("GetRSAPrivateKey， block == nil")
		return nil
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("GetRSAPrivateKey， x509.ParsePKCS8PrivateKey 失败: %+v", err)
		return nil
	}

	return privateKey
}

func (s *SecurityConfig) GetRSAPublicKey() interface{} {
	log.Infof("GetRSAPublicKey， RSAPrivateKey: %s", s.RSAPrivateKey)
	block, _ := pem.Decode([]byte(s.RSAPrivateKey))
	if block == nil {
		log.Errorf("GetRSAPublicKey， block == nil")
		return nil
	}

	// 先解析私钥
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("GetRSAPublicKey， 解析私钥失败: %+v", err)
		return nil
	}

	// 根据私钥类型获取公钥
	switch pk := privateKey.(type) {
	case *rsa.PrivateKey:
		return &pk.PublicKey
	default:
		log.Errorf("GetRSAPublicKey， 不支持的私钥类型")
		return nil
	}
}

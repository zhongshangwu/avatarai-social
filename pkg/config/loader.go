package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*SocialConfig, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigName("default")
	v.SetConfigType("yaml")
	v.AddConfigPath("./conf")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取默认配置文件失败: %w", err)
	}

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("合并配置文件失败: %w", err)
		}
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("AVATARAI")

	var config SocialConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.http.address", ":8080")
	v.SetDefault("server.http.read_timeout", "30s")
	v.SetDefault("server.http.write_timeout", "30s")
	v.SetDefault("server.http.idle_timeout", "60s")

	v.SetDefault("server.https.enabled", false)
	v.SetDefault("server.https.address", ":8443")
	v.SetDefault("server.https.cert_file", "certs/server.crt")
	v.SetDefault("server.https.key_file", "certs/server.key")
	v.SetDefault("server.https.read_timeout", "30s")
	v.SetDefault("server.https.write_timeout", "30s")
	v.SetDefault("server.https.idle_timeout", "60s")

	v.SetDefault("server.metrics.address", ":8081")

	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "./data/avatarai.sqlite")
	v.SetDefault("database.max_connections", 20)
	v.SetDefault("database.max_idle_connections", 5)
	v.SetDefault("database.connection_lifetime", "1h")
	v.SetDefault("database.connection_timeout", "30s")

	v.SetDefault("storage.data_dir", "data/avatarai")

	v.SetDefault("avatar.llm.api_url", "https://api.openai.com/v1")
	v.SetDefault("avatar.llm.model", "gpt-4")
	v.SetDefault("avatar.llm.provider", "openai")
}

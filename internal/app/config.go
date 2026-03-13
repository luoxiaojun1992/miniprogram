package app

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Wechat   WechatConfig   `mapstructure:"wechat"`
	Upload   UploadConfig   `mapstructure:"upload"`
	Debug    DebugConfig    `mapstructure:"debug"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expiry int    `mapstructure:"expiry"`
}

// LogConfig holds log configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// WechatConfig holds wechat configuration.
type WechatConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

// UploadConfig holds upload configuration.
type UploadConfig struct {
	Dir     string `mapstructure:"dir"`
	BaseURL string `mapstructure:"base_url"`
}

// DebugConfig holds debug/development configuration.
// These settings MUST NOT be enabled in production.
type DebugConfig struct {
	// EnableTestToken enables the POST /v1/debug/token endpoint that issues
	// JWT tokens without authentication — for local testing only.
	// Default: false
	EnableTestToken bool `mapstructure:"enable_test_token"`
}

// InitConfig loads configuration using Viper.
// If configPath is empty the configuration is read entirely from environment
// variables (prefix APP_).  When configPath is provided the file is loaded
// first and environment variables override individual keys.
func InitConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// Defaults make Unmarshal work even when no config file is present.
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "miniapp")
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.expiry", 7200)
	v.SetDefault("log.level", "info")
	v.SetDefault("wechat.app_id", "")
	v.SetDefault("wechat.app_secret", "")
	v.SetDefault("upload.dir", "storage/uploads")
	v.SetDefault("upload.base_url", "")
	v.SetDefault("debug.enable_test_token", false)

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

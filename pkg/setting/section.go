// Package setting contains the configuration settings for the application.
package setting

type Config struct {
	Server    Server    `mapstructure:"server"`
	Mongo     Mongo     `mapstructure:"mongo"`
	MySQL     MySQL     `mapstructure:"mysql"`
	Redis     Redis     `mapstructure:"redis"`
	Security  Security  `mapstructure:"security"`
	Logger    Logger    `mapstructure:"logger"`
	Minio     Minio     `mapstructure:"minio"`
	RateLimit RateLimit `mapstructure:"ratelimit"`
	Cors      Cors      `mapstructure:"cors"`
}

// Server config
type Server struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type Mongo struct {
	URI      string `mapstructure:"uri"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"dbName"`
}

// MySQL config
type MySQL struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"maxIdleConns"`
	MaxOpenConns    int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime int    `mapstructure:"connMaxLifetime"`
}

// Redis config
type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	Pool     int    `mapstructure:"pool"`
}

// Security config
type Security struct {
	JWT struct {
		AccessTokenSecret       string `mapstructure:"access-secret"`
		AccessTokenExpiration   string `mapstructure:"access-expiration"`
		RefreshTokenSecret      string `mapstructure:"refresh-secret"`
		RefreshTokenExpiration  string `mapstructure:"refresh-expiration"`
		ResetPasswordSecret     string `mapstructure:"reset-password-secret"`
		ResetPasswordExpiration string `mapstructure:"reset-password-expiration"`
	}
}

type Logger struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge"`
	Compress   bool   `mapstructure:"compress"`
}

type RateLimit struct {
	Limit    int    `mapstructure:"limit"`
	Interval string `mapstructure:"interval"` // e.g., "1s", "10s", "1m"
	Burst    int    `mapstructure:"burst"`
}

type Minio struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accessKey"`
	SecretKey string `mapstructure:"secretKey"`
	UseSSL    bool   `mapstructure:"useSSL"`
}

type Cors struct {
	AllowOrigins []string `mapstructure:"allowOrigins"`
}

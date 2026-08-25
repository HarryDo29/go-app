// Package setting contains the configuration settings for the application.
package setting

type Config struct {
	Server    Server    `mapstructure:"server"`
	Mongo     Mongo     `mapstructure:"mongo"`
	MySQL     MySQL     `mapstructure:"mysql"`
	Redis     Redis     `mapstructure:"redis"`
	Security  Security  `mapstructure:"security"`
	Logger    Logger    `mapstructure:"logger"`
	S3        S3        `mapstructure:"s3"`
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
	MaxIdleConns    int64  `mapstructure:"maxIdleConns"`
	MaxOpenConns    int64  `mapstructure:"maxOpenConns"`
	ConnMaxLifetime int64  `mapstructure:"connMaxLifetime"`
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

type S3 struct {
	Provider      string `mapstructure:"provider"`      // "minio" or "aws"
	Endpoint      string `mapstructure:"endpoint"`      // e.g. "http://localhost:9000" for MinIO; leave empty for AWS S3
	Region        string `mapstructure:"region"`        // e.g. "us-east-1" or "ap-southeast-1"
	BucketName    string `mapstructure:"bucketName"`    // e.g. "go-chat-app"
	AccessKey     string `mapstructure:"accessKey"`     // Access Key
	SecretKey     string `mapstructure:"secretKey"`     // Secret Key
	UseSSL        bool   `mapstructure:"useSSL"`        // SSL enabled
	UsePathStyle  bool   `mapstructure:"usePathStyle"`  // true for MinIO, false for AWS S3
	MaxSingleSize int64  `mapstructure:"maxSingleSize"` // 5MB in bytes
	MinPartSize   int64  `mapstructure:"minPartSize"`   // 5MB in bytes (S3 minimum)
	MaxFileSize   int64  `mapstructure:"maxFileSize"`   // Max total file size in bytes

}

type Cors struct {
	AllowOrigins []string `mapstructure:"allowOrigins"`
}

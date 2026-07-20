package initianlize

import (
	"fmt"
	"go-app/global"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	Databases []struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
	} `mapstructure:"databases"`
	Security struct {
		JWT struct {
			Secret     string `mapstructure:"secret"`
			Expiration string `mapstructure:"expiration"`
		} `mapstructure:"jwt"`
	} `mapstructure:"security"`
}

func LoadConfig() {
	v := viper.New()
	v.AddConfigPath("./config")
	v.SetConfigName("local")
	v.SetConfigType("yaml")

	// Đọc biến môi trường (Environment Variables)
	v.AutomaticEnv()
	// Map dấu . sang _ (vd: server.port -> SERVER_PORT)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// read configuration
	err := v.ReadInConfig()
	if err != nil {
		// Bỏ qua lỗi nếu không tìm thấy file config (khi chạy trên docker/portainer)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("error reading config: %s\n", err))
		}
	}

	if err = v.Unmarshal(&global.Config); err != nil {
		fmt.Printf("error unmarshalling config: %s\n", err)
	}
}

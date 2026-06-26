package initianlize

import (
	"fmt"
	"go-app/global"

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
	viper := viper.New()
	viper.AddConfigPath("./config")
	viper.SetConfigName("local")
	viper.SetConfigType("yaml")

	// read configuration
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("error reading config: %s\n", err))
	}

	// read data
	// fmt.Printf("Server port: %d\n", viper.GetInt("server.port"))

	if err = viper.Unmarshal(&global.Config); err != nil {
		fmt.Printf("error unmarshalling config: %s\n", err)
	}

	// fmt.Println("Config Port: ", global.Config.Server.Port)
}

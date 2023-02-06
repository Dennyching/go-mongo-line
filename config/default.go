package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUri                  string `mapstructure:"MONGODB_LOCAL_URI"`
	RedisUri               string `mapstructure:"REDIS_URL"`
	Port                   string `mapstructure:"PORT"`
	LineChannelAccessToken string `mapstructure:"LINE_BOT_CHANNEL_ACCESS_TOKEN"`
	LineChannelSecret      string `mapstructure:"LINE_BOT_CHANNEL_SECRET"`
	LineBotUserId          string `mapstructure:"LINE_BOT_USER_ID"`
}

// ? Struct Config

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

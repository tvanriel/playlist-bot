package config

import (
	"strings"

	"github.com/mitaka8/playlist-bot/internal/discord"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/youtubedl"
	"github.com/spf13/viper"
	"github.com/tvanriel/cloudsdk/amqp"
	"github.com/tvanriel/cloudsdk/http"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"github.com/tvanriel/cloudsdk/logging"
	"github.com/tvanriel/cloudsdk/mysql"
	"github.com/tvanriel/cloudsdk/redis"
	"github.com/tvanriel/cloudsdk/s3"
)

type Configuration struct {
	Http       http.Configuration       `mapstructure:"http"`
	Amqp       amqp.Configuration       `mapstructure:"amqp"`
	Kubernetes kubernetes.Configuration `mapstructure:"kubernetes"`
	S3         s3.Configuration         `mapstructure:"s3"`
	Mysql      mysql.Configuration      `mapstructure:"mysql"`
	Logging    logging.Configuration    `mapstructure:"log"`
	Discord    discord.Configuration    `mapstructure:"discord"`
	YouTubeDL  youtubedl.Configuration  `mapstructure:"youtube"`
	MusicStore musicstore.Configuration `mapstructure:"musicstore"`
	Redis      redis.Configuration      `mapstructure:"redis"`
}

func ViperConfiguration() (*Configuration, error) {
	config := &Configuration{}
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/discordbot")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return config, err

}

func HttpConfiguration(config *Configuration) http.Configuration {
	return config.Http
}

func AmqpConfiguration(config *Configuration) amqp.Configuration {
	return config.Amqp
}

func KubernetesConfiguration(config *Configuration) *kubernetes.Configuration {
	return &config.Kubernetes
}

func MySQLConfiguration(config *Configuration) mysql.Configuration {
	return config.Mysql
}

func LoggingConfiguration(config *Configuration) logging.Configuration {
	return config.Logging
}

func S3Configuration(config *Configuration) *s3.Configuration {
	return &config.S3
}
func DiscordConfiguration(config *Configuration) *discord.Configuration {
	return &config.Discord
}
func YoutubeDLConfiguration(config *Configuration) youtubedl.Configuration {
	return config.YouTubeDL
}

func MusicStoreConfiguration(config *Configuration) *musicstore.Configuration {
	return &config.MusicStore
}
func RedisConfiguration(config *Configuration) redis.Configuration {
	return config.Redis

}

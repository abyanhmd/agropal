package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Stage  string `yaml:"stage"`
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Path struct {
		Migration string `yaml:"migration"`
		Log       string `yaml:"log"`
	} `yaml:"path"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Dbname   string `mapstructure:"dbname"`
	} `mapstructure:"postgres"`
	Redis struct {
		Address  string `mapstructure:"address"`
		Password string `mapstructure:"password"`
		Db       int    `mapstructure:"db"`
		Protocol int    `mapstructure:"protocol"`
	} `mapstructure:"redis"`
	RabbitMQ struct {
		Address  string `mapstructure:"address"`
		URL      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"rabbitmq"`
	Smtp struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Username string `mapstructure:"username"`
		From     string `mapstructure:"from"`
	} `mapstructure:"smtp"`
	Secret struct {
		Jwt      string `mapstructure:"jwt"`
		GcmToken string `mapstructure:"gcm_token"`
		Smtp     string `mapstructure:"smtp"`
	} `mapstructure:"secret"`
}

func LoadConfig(path string) *Config {
	if _, exists := os.LookupEnv("ENV"); !exists {
		if err := godotenv.Load(".env"); err != nil {
			panic(err)
		}
	}

	v := viper.New()
	v.SetConfigFile(path)

	bindings := map[string]string{
		"postgres.host":     "POSTGRES_HOST",
		"postgres.port":     "POSTGRES_PORT",
		"postgres.user":     "POSTGRES_USER",
		"postgres.password": "POSTGRES_PASSWORD",
		"postgres.dbname":   "POSTGRES_NAME",

		"redis.address":  "REDIS_ADDRESS",
		"redis.password": "REDIS_PASSWORD",
		"redis.db":       "REDIS_DB",
		"redis.protocol": "REDIS_PROTOCOL",

		"rabbitmq.address":  "RABBITMQ_ADDRESS",
		"rabbitmq.url":      "RABBITMQ_URL",
		"rabbitmq.username": "RABBITMQ_USERNAME",
		"rabbitmq.password": "RABBITMQ_PASSWORD",

		"smtp.host":     "SMTP_HOST",
		"smtp.port":     "SMTP_PORT",
		"smtp.username": "SMTP_USERNAME",
		"smtp.from":     "SMTP_FROM",

		"secret.jwt":       "SECRET_JWT",
		"secret.gcm_token": "SECRET_GCM_TOKEN",
		"secret.smtp":      "SECRET_SMTP",
	}

	for key, env := range bindings {
		if err := v.BindEnv(key, env); err != nil {
			panic(err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}

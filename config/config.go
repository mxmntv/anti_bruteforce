package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/mxmntv/anti_bruteforce/internal/model"
)

type Config struct {
	App                  `yaml:"app"`
	HTTP                 `yaml:"http"`
	Log                  `yaml:"logger"`
	Redis                `yaml:"redis"`
	model.BucketCapacity `yaml:"capacity"`
}

type App struct {
	Name       string `env-required:"true" yaml:"name"    env:"APP_NAME"`
	Version    string `env-required:"true" yaml:"version"    env:"APP_VERSION"`
	APIVersion string `env-required:"false" yaml:"apiVersion"    env:"API_VERSION"`
}

type HTTP struct {
	Port int `env-required:"true" yaml:"port"    env:"HTTP_PORT"`
}

type Log struct {
	Level string `env-required:"true" yaml:"logLevel"    env:"LOG_LEVEL"`
}

type Redis struct {
	RsHost string `env-required:"true" yaml:"redisHost"    env:"RS_HOST"`
	RsPort int    `env-required:"true" yaml:"redisPort"    env:"RS_PORT"`
}

func NewConfig(path string) (*Config, error) {
	config := &Config{}
	err := cleanenv.ReadConfig(path, config)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

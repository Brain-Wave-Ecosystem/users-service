package config

import "github.com/Brain-Wave-Ecosystem/go-common/pkg/config"

type Config struct {
	config.DefaultServiceConfig
	Redis    RedisConfig    `envPrefix:"REDIS_"`
	Postgres PostgresConfig `envPrefix:"POSTGRES_"`
}

type RedisConfig struct {
	URL string `env:"URL"`
}

type PostgresConfig struct {
	URL string `env:"URL"`
}

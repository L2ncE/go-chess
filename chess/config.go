package main

type ServerConfig struct {
	Name      string      `mapstructure:"name"`
	Port      int         `mapstructure:"port"`
	RedisInfo RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

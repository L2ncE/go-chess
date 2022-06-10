package model

type ServerConfig struct {
	Name      string      `mapstructure:"name"`
	Port      int         `mapstructure:"port"`
	GormInfo  GormConfig  `mapstructure:"gorm"`
	RedisInfo RedisConfig `mapstructure:"redis"`
}

type GormConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbName"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

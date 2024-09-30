package config

type Config struct {
	DBConfig *DBConfig
	RedisConfig *RedisConfig
}

func LoadConfig() *Config {
	cfg := &Config{
		DBConfig: loadDBConfig(),
		RedisConfig: loadRedisConfig(),
	}
	return cfg
}

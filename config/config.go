package config

type Config struct {
	DBConfig      *DBConfig
}


func LoadConfig() *Config {
	cfg := &Config{
		DBConfig: loadDBConfig(),
	}
	return cfg
}
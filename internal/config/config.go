package config

type Config struct {
	RtmpServerPort int
	HttpServerAddr string
}

func NewConfig() *Config {
	return &Config{
		RtmpServerPort: 1935,
		HttpServerAddr: ":8000",
	}
}

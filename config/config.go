package config

import (
	env "github.com/Netflix/go-env"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Logging struct {
		Level     string `env:"LOG_LEVEL,default=info"`
		Formatter string `env:"LOG_FORMATTER,default=json"`
	}
	CollectionInterval string `env:"COLLECT_INTERVAL,default=1h"`
	UploadInventory    bool   `env:"UPLOAD_INVENTORY,default=true"`
	Impersonate        string `env:"IMPERSONATE"`
	ServerAPIEndpoint  string `env:"SERVER_API_ENDPOINT,default=http://localhost:8086"`
	HTTPPort           string `env:"HTTP_PORT,default=8087"`
	TLSCrt             string `env:"TLS_CRT,default=/etc/certificates/tls.crt"`
	TLSKey             string `env:"TLS_KEY,default=/etc/certificates/tls.key"`
	AuthEnabled        bool   `env:"AUTH_ENABLED,default=true"`
	Debug              bool   `env:"DEBUG,default=false"`
	Extras             env.EnvSet
}

func NewConfig() Config {
	var c Config
	es, err := env.UnmarshalFromEnviron(&c)
	if err != nil {
		log.Fatal().Err(err).Msg("getting environment variables")
	}
	c.Extras = es
	return c
}

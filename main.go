package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/neticdk-k8s/k8s-inventory-client/collect"
	"github.com/neticdk-k8s/k8s-inventory-client/collect/version"
	"github.com/neticdk-k8s/k8s-inventory-client/config"
	"github.com/neticdk-k8s/k8s-inventory-client/logging"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.NewConfig()
	logging.InitLogger(cfg.Logging.Level, cfg.Logging.Formatter)

	log.Info().Str("version", version.VERSION).Str("commit", version.COMMIT).Msg("starting k8s-inventory-client")

	collection := collect.NewInventoryCollection(cfg)

	go collection.Collect()

	http.Handle("/api/v1/inventory", collection)

	log.Info().Str("port", cfg.HTTPPort).Msg("starting server")
	err := http.ListenAndServe(":"+cfg.HTTPPort, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

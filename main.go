package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/neticdk-k8s/k8s-inventory-client/collect"
	"github.com/neticdk-k8s/k8s-inventory-client/collect/version"
	"github.com/neticdk-k8s/k8s-inventory-client/config"
	"github.com/neticdk-k8s/k8s-inventory-client/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pkg/profile"
)

func main() {
	cfg := config.NewConfig()
	logging.InitLogger(cfg.Logging.Level, cfg.Logging.Formatter)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debugging mode enabled")
		log.Debug().Msg("Starting profiler")
		defer profile.Start(profile.MemProfile).Stop()
		go func() {
			if err := http.ListenAndServe(":8080", nil); err != nil {
				panic(err)
			}
		}()
	}

	log.Info().Str("version", version.VERSION).Str("commit", version.COMMIT).Msg("starting k8s-inventory-client")

	collection := collect.NewInventoryCollection(cfg)

	go collection.Collect()

	http.Handle("/", collection)

	metaHandler := http.NewServeMux()
	metaHandler.HandleFunc("/", collection.ServeHTTPMeta)

	go func() {
		log.Info().Str("portMeta", cfg.HTTPPortMeta).Msg("starting metadata server")
		if err := http.ListenAndServe(":"+cfg.HTTPPortMeta, metaHandler); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}()

	log.Info().Str("port", cfg.HTTPPort).Msg("starting server")
	if err := http.ListenAndServe(":"+cfg.HTTPPort, nil); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

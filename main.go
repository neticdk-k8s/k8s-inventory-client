package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/neticdk-k8s/k8s-inventory-client/collect"
	"github.com/neticdk-k8s/k8s-inventory-client/logging"
	"github.com/rs/zerolog/log"
)

var (
	VERSION = "dev"
	COMMIT  = "HEAD"
)

func getDefaultValue(envVar, def string) string {
	r, ok := os.LookupEnv(envVar)
	if !ok {
		return def
	}
	return r
}

func main() {
	logLevel := getDefaultValue("LOG_LEVEL", "info")
	logFormatter := getDefaultValue("LOG_FORMATTER", "json")
	logging.InitLogger(logLevel, logFormatter)

	log.Info().Str("version", VERSION).Str("commit", COMMIT).Msg("starting k8s-inventory-client")

	collectionInterval := getDefaultValue("COLLECT_INTERVAL", collect.DefaultCollectionInterval)
	uploadInventory := getDefaultValue("UPLOAD_INVENTORY", "true")
	impersonate := getDefaultValue("IMPERSONATE", "")
	serverAPIEndpoint := getDefaultValue("SERVER_API_ENDPOINT", "http://localhost:8086")
	collection := collect.NewInventoryCollection(collectionInterval, uploadInventory, impersonate, serverAPIEndpoint)

	go collection.Collect()

	http.Handle("/api/v1/inventory", collection)

	port := getDefaultValue("HTTP_PORT", "8087")
	log.Info().Str("port", port).Msg("starting server")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

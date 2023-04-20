package main

import (
	"net/http"
	"os"

	"github.com/neticdk-k8s/k8s-inventory-client/collect"
	"github.com/neticdk-k8s/k8s-inventory-client/logging"
	log "github.com/sirupsen/logrus"
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

	log.Infof("Starting k8s-inventory-client, version: %s, commit: %s", VERSION, COMMIT)

	collectionInterval := getDefaultValue("COLLECT_INTERVAL", collect.DefaultCollectionInterval)
	uploadInventory := getDefaultValue("UPLOAD_INVENTORY", "true")
	impersonate := getDefaultValue("IMPERSONATE", "")
	serverAPIEndpoint := getDefaultValue("SERVER_API_ENDPOINT", "http://localhost:8086")
	collection := collect.NewInventoryCollection(collectionInterval, uploadInventory, impersonate, serverAPIEndpoint)
	go collection.Collect()

	http.Handle("/api/v1/inventory", collection)

	port := getDefaultValue("HTTP_PORT", "8087")
	log.Infof("Serving on port %v", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("An error occured: %s", err)
	}
}

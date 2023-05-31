package collect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	kubernetes "github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// How often to collect
const DefaultCollectionInterval = "1h"

type InventoryCollection struct {
	Inventory          *inventory.Inventory
	CollectionInterval string
	UploadInventory    bool
	Impersonate        string
	ServerAPIEndpoint  string
}

func NewInventoryCollection(collectionInterval string, uploadInventory string, impersonate string, serverAPIEndpoint string) *InventoryCollection {
	return &InventoryCollection{
		CollectionInterval: collectionInterval,
		UploadInventory:    uploadInventory == "true",
		Impersonate:        impersonate,
		ServerAPIEndpoint:  fmt.Sprintf("%s/api/v1/inventory", serverAPIEndpoint),
	}
}

func (c *InventoryCollection) Collect() {
	r, err := time.ParseDuration(c.CollectionInterval)
	if err != nil {
		log.Warn().Err(err).Str("interval", c.CollectionInterval).Msg("parsing refresh interval")
		r, err = time.ParseDuration(DefaultCollectionInterval)
		if err != nil {
			log.Fatal().Err(err).Str("interval", DefaultCollectionInterval).Msg("parsing refresh interval")
		}
	}

	sleepNext := func() {
		t := time.Now().Add(r)
		log.Info().Msgf("next iteration in %v at %v", r, t.Local().Format("2006-01-02 15:04:05"))
		time.Sleep(r)
	}

	log.Info().Msg("entering inventory collection loop")
	for {
		c.Inventory = inventory.NewInventory()
		c.Inventory.CollectionSucceeded = true
		cs, err := kubernetes.CreateK8SClient(c.Impersonate)
		if err != nil {
			log.Error().Err(err).Msg("creating clientset")
			c.Inventory.CollectionSucceeded = false
			sleepNext()
			continue
		}

		log.Debug().Str("collect", "cluster").Msg("")
		c.handleErrors(CollectCluster(cs, c.Inventory))

		log.Debug().Str("collect", "scs").Msg("")
		c.handleErrors(CollectSCSMetadata(cs, c.Inventory))
		c.handleErrors(CollectSCSTenants(cs, c.Inventory))

		log.Debug().Str("collect", "namespace").Msg("")
		c.handleErrors(CollectNamespaces(cs, c.Inventory))

		log.Debug().Str("collect", "node").Msg("")
		c.handleErrors(CollectNodes(cs, c.Inventory))

		log.Debug().Str("collect", "storage").Msg("")
		c.handleErrors(CollectStorage(cs, c.Inventory))

		log.Debug().Str("collect", "components").Msg("")
		c.handleErrors(CollectCustomResources(cs, c.Inventory))

		log.Debug().Str("collect", "workload").Msg("")
		c.handleErrors(CollectWorkloads(cs, c.Inventory))

		if c.UploadInventory {
			if err := c.Upload(); err != nil {
				log.Error().Stack().Err(err).Msg("uplading inventory")
			}
		}

		sleepNext()
	}
}

func (c *InventoryCollection) Upload() error {
	log.Info().Msg("uploading inventory")

	payload, err := json.Marshal(c.Inventory)
	if err != nil {
		return errors.Wrap(err, "marshaling inventory")
	}

	req, err := http.NewRequest("PUT", c.ServerAPIEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "sending request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Error().Err(err).Int("status", res.StatusCode).Msg("reading response")
			return errors.Wrap(err, "reading response")
		}
		log.Error().Int("status", res.StatusCode).Str("body", string(body)).Msg("upload failed")
	}

	log.Info().Str("fqdn", c.Inventory.Cluster.Name).Int("status", res.StatusCode).Msg("uploaded inventory")

	return nil
}

func (c *InventoryCollection) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	viJSON, err := json.Marshal(c.Inventory)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(viJSON)
	if err != nil {
		panic(err)
	}
}

func (c *InventoryCollection) handleErrors(errs []error) {
	if len(errs) > 0 {
		c.Inventory.CollectionSucceeded = false
	}
	for _, e := range errs {
		log.Error().Stack().Err(e).Msg("")
	}
}

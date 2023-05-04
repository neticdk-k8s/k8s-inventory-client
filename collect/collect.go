package collect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	kubernetes "github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
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
		Inventory:          inventory.NewInventory(),
		CollectionInterval: collectionInterval,
		UploadInventory:    uploadInventory == "true",
		Impersonate:        impersonate,
		ServerAPIEndpoint:  serverAPIEndpoint,
	}
}

func (c *InventoryCollection) Collect() {
	r, err := time.ParseDuration(c.CollectionInterval)
	if err != nil {
		log.Warnf("parsing refresh interval: %v. Using default: %v",
			err, DefaultCollectionInterval)
		r, err = time.ParseDuration(DefaultCollectionInterval)
		if err != nil {
			log.Fatalf("parsing refresh interval: %v", err)
		}
	}

	sleepNext := func() {
		t := time.Now().Add(r)
		log.Infof("Next iteration in %v at %v", r, t.Local().Format("2006-01-02 15:04:05"))
		time.Sleep(r)

	}

	log.Infof("Entering inventory collection loop")
	for {
		c.Inventory.CollectionSucceeded = true
		cs, err := kubernetes.CreateK8SClient(c.Impersonate)
		if err != nil {
			log.Errorf("creating clientset: %v", err)
			c.Inventory.CollectionSucceeded = false
			sleepNext()
			continue
		}

		log.Infof("Collecting cluster information")
		c.handleErrors(CollectCluster(cs, c.Inventory))

		log.Infof("Collecting Secure Cloud Stack information")
		c.handleErrors(CollectSCSMetadata(cs, c.Inventory))
		c.handleErrors(CollectSCSTenants(cs, c.Inventory))

		log.Infof("Collecting namespace information")
		c.handleErrors(CollectNamespaces(cs, c.Inventory))

		log.Infof("Collecting node information")
		c.handleErrors(CollectNodes(cs, c.Inventory))

		log.Infof("Collecting storage information")
		c.handleErrors(CollectStorage(cs, c.Inventory))

		log.Infof("Collecting custom resources information")
		c.handleErrors(CollectCustomResources(cs, c.Inventory))

		log.Infof("Collecting workload information")
		c.handleErrors(CollectWorkloads(cs, c.Inventory))

		if c.UploadInventory {
			c.Upload()
		}

		sleepNext()
	}
}

func (c *InventoryCollection) Upload() {
	log.Infof("Uploading inventory")
	serverAPIEndpoint := fmt.Sprintf("%s/api/v1/inventory", c.ServerAPIEndpoint)

	payload, err := json.Marshal(c.Inventory)
	if err != nil {
		log.Error(err)
		return
	}

	req, err := http.NewRequest("PUT", serverAPIEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		log.Error(err)
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		log.WithFields(log.Fields{
			"status": res.StatusCode,
			"body":   string(body),
		}).Error("upload failed")
		return
	}

	log.Infof("Uploaded inventory for: %s (%d)", c.Inventory.Cluster.Name, res.StatusCode)
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
		log.Error(e)
	}
}

func filterAnnotations(src metav1.Object) (a map[string]string) {
	a = src.GetAnnotations()
	if len(a) > 0 {
		delete(a, "kubectl.kubernetes.io/last-applied-configuration")
	} else {
		a = make(map[string]string)
	}
	return a
}

func readConfigMapsByLabel(cs *ck.Clientset, ns string, label string) (data []map[string]string, err error) {
	data = make([]map[string]string, 0)
	res, err := cs.CoreV1().
		ConfigMaps(ns).
		List(context.TODO(), metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return data, err
	}
	for i := range res.Items {
		data = append(data, res.Items[i].Data)
	}
	return data, nil
}

func appendError(dst []error, errs ...error) []error {
	for _, e := range errs {
		if e != nil {
			dst = append(dst, e)
		}
	}
	return dst
}

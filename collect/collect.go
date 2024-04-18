package collect

import (
	"bytes"
	"compress/gzip"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/collect/version"
	"github.com/neticdk-k8s/k8s-inventory-client/config"
	kubernetes "github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	jose "gopkg.in/go-jose/go-jose.v2"
)

// How often to collect
const defaultCollectionInterval = "1h"

type InventoryCollection struct {
	mu                 sync.RWMutex
	inventory          *inventory.Inventory
	collectionInterval string
	uploadInventory    bool
	impersonate        string
	serverAPIEndpoint  string
	tlsCrt             string
	tlsKey             string
	authEnabled        bool
	signer             jose.Signer
	metaData           *metaData
}

type metaData struct {
	Updated  *time.Time              `json:"updated,omitempty"`
	Cluster  *uploadResponseCluster  `json:"cluster,omitempty"`
	MetaData *uploadResponseMetaData `json:"meta_data,omitempty"`
}

type uploadResponseCluster struct {
	Name         string `json:"name,omitempty"`
	OperatorName string `json:"operator_name,omitempty"`
	ProviderName string `json:"provider_name,omitempty"`
	ID           string `json:"id,omitempty"`
}

type uploadResponseServiceLevel struct {
	HasTechnicalOperations   *bool  `json:"has_technical_operations,omitempty"`
	HasTechnicalManagement   *bool  `json:"has_technical_management,omitempty"`
	HasApplicationOperations *bool  `json:"has_application_operations,omitempty"`
	HasApplicationManagement *bool  `json:"has_application_management,omitempty"`
	HasCustomOperations      *bool  `json:"has_custom_operations,omitempty"`
	CustomOperationsURL      string `json:"custom_operations_url,omitempty"`
}

type uploadResponseMetaData struct {
	ClusterType            string                     `json:"cluster_type,omitempty"`
	Description            string                     `json:"description,omitempty"`
	Partition              string                     `json:"partition,omitempty"`
	Region                 string                     `json:"region,omitempty"`
	EnvironmentName        string                     `json:"environment_name,omitempty"`
	InfrastructureProvider string                     `json:"infrastructure_provider,omitempty"`
	ResilienceZone         string                     `json:"resilience_zone,omitempty"`
	ServiceLevel           uploadResponseServiceLevel `json:"service_level,omitempty"`
	SubscriptionID         string                     `json:"subscription_id,omitempty"`
}

type uploadResponse struct {
	Cluster  uploadResponseCluster  `json:"cluster"`
	MetaData uploadResponseMetaData `json:"meta_data"`
	Message  string                 `json:"message,omitempty"`
}

func NewInventoryCollection(cfg config.Config) *InventoryCollection {
	i := &InventoryCollection{
		collectionInterval: cfg.CollectionInterval,
		uploadInventory:    cfg.UploadInventory,
		impersonate:        cfg.Impersonate,
		serverAPIEndpoint:  fmt.Sprintf("%s/api/v1/inventory", cfg.ServerAPIEndpoint),
		tlsCrt:             cfg.TLSCrt,
		tlsKey:             cfg.TLSKey,
		authEnabled:        cfg.AuthEnabled,
		metaData:           &metaData{},
	}
	if !i.authEnabled {
		log.Info().Msg("Authentication disabled")
		return i
	}
	if i.tlsCrt == "" {
		log.Info().Msg("No TLSCrt set. Authentication disabled")
		i.authEnabled = false
		return i
	}
	if i.tlsKey == "" {
		log.Error().Msg("No TLSKey set. Authentication disbaled.")
		i.authEnabled = false
		return i
	}
	if _, err := os.Stat(filepath.Clean(i.tlsCrt)); errors.Is(err, os.ErrNotExist) {
		log.Error().Msgf("TLSCrt file '%s' not found. Authentication disbaled.", i.tlsCrt)
		i.authEnabled = false
		return i
	}
	if _, err := os.Stat(filepath.Clean(i.tlsKey)); errors.Is(err, os.ErrNotExist) {
		log.Error().Msgf("TLSKey file '%s' not found. Authentication disbaled.", i.tlsKey)
		i.authEnabled = false
		return i
	}
	log.Info().Msg("Authentication enabled")
	i.authEnabled = true
	i.refreshCertificates()
	return i
}

func (c *InventoryCollection) refreshCertificates() {
	reschedule := func(d time.Duration) {
		log.Info().Dur("duration", d).Msg("refreshing key and certificate")
		t := time.NewTimer(d)
		<-t.C
		c.refreshCertificates()
	}

	certificates, key, err := readKeyAndCertificates(c.tlsCrt, c.tlsKey)
	if err != nil {
		log.Error().Err(err).Msg("unable to read certificates")
		go reschedule(2 * time.Minute)
		return
	}

	jwk := jose.JSONWebKey{Certificates: certificates, Key: key}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS512, Key: jwk}, &jose.SignerOptions{EmbedJWK: true})
	if err != nil {
		log.Error().Err(err).Msg("unable to create JOSE signer")
		go reschedule(2 * time.Minute)
		return
	}
	c.signer = signer

	if len(certificates) > 0 {
		d := time.Until(certificates[0].NotAfter)
		go reschedule(d - (5 * time.Minute))
	}
}

func readKeyAndCertificates(certFile, keyFile string) ([]*x509.Certificate, any, error) {
	certificates := []*x509.Certificate{}
	data, err := os.ReadFile(filepath.Clean(certFile))
	if err != nil {
		return nil, nil, err
	}
	for len(data) > 0 {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, nil, err
		}
		certificates = append(certificates, cert)
	}

	data, err = os.ReadFile(filepath.Clean(keyFile))
	if err != nil {
		return nil, nil, err
	}
	p, _ := pem.Decode(data)
	key, err := x509.ParsePKCS8PrivateKey(p.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(p.Bytes)
		if err != nil {
			return nil, nil, err
		}
	}

	return certificates, key, nil
}

func (c *InventoryCollection) Collect() {
	r, err := time.ParseDuration(c.collectionInterval)
	if err != nil {
		log.Warn().Err(err).Str("interval", c.collectionInterval).Msg("parsing refresh interval")
		r, err = time.ParseDuration(defaultCollectionInterval)
		if err != nil {
			log.Fatal().Err(err).Str("interval", defaultCollectionInterval).Msg("parsing refresh interval")
		}
	}

	sleepNext := func() {
		t := time.Now().Add(r)
		log.Info().Msgf("next iteration in %v at %v", r, t.Local().Format(time.DateTime))
		time.Sleep(r)
	}

	log.Info().Msg("entering inventory collection loop")
	for {
		c.inventory = inventory.NewInventory()
		c.inventory.CollectionSucceeded = true
		c.inventory.ClientVersion = version.VERSION
		c.inventory.ClientCommit = version.COMMIT
		cs, client, err := kubernetes.CreateK8SClient(c.impersonate)
		if err != nil {
			log.Error().Err(err).Msg("creating clientset")
			c.inventory.CollectionSucceeded = false
			c.inventory.CollectionErrors = append(c.inventory.CollectionErrors, err.Error())
			sleepNext()
			continue
		}

		log.Debug().Str("collect", "cluster").Msg("")
		c.handleError(collectCluster(cs, c.inventory))

		log.Debug().Str("collect", "scs").Msg("")
		c.handleError(collectSCSMetadata(cs, c.inventory))

		log.Debug().Str("collect", "namespace").Msg("")
		c.handleError(collectNamespaces(cs, c.inventory))

		log.Debug().Str("collect", "node").Msg("")
		c.handleError(collectNodes(cs, c.inventory))

		log.Debug().Str("collect", "storage").Msg("")
		c.handleError(collectStorage(cs, c.inventory))

		log.Debug().Str("collect", "network_policy").Msg("")
		c.handleError(collectNetworkPolicies(cs, c.inventory))

		log.Debug().Str("collect", "components").Msg("")
		c.handleError(collectCustomResources(cs, c.inventory))

		log.Debug().Str("collect", "workload").Msg("")
		c.handleError(collectWorkloads(cs, client, c.inventory))

		if c.uploadInventory {
			if err := c.upload(); err != nil {
				log.Error().Stack().Err(err).Msg("uplading inventory")
			}
		}

		sleepNext()
	}
}

func (c *InventoryCollection) upload() error {
	log.Info().Msg("uploading inventory")

	var (
		payload     []byte
		err         error
		contentType string = "application/json; charset=UTF-8"
	)

	payload, err = json.Marshal(c.inventory)
	if err != nil {
		return errors.Wrap(err, "marshaling inventory")
	}

	if c.authEnabled {
		jws, err := c.signer.Sign(payload)
		if err != nil {
			return errors.Wrap(err, "signing inventory")
		}
		payload = []byte(jws.FullSerialize())
		contentType = "application/jose+json"
	}

	var gzippedBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzippedBuf)
	if _, err = gzipWriter.Write(bytes.NewBuffer(payload).Bytes()); err != nil {
		return errors.Wrap(err, "compressiong payload")
	}
	if err = gzipWriter.Close(); err != nil {
		return errors.Wrap(err, "compressiong payload")
	}

	req, err := http.NewRequest("PUT", c.serverAPIEndpoint, &gzippedBuf)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")
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
		log.Error().Int("status", res.StatusCode).Str("body", string(body)).Msg("")
		return errors.New("upload failed")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	metaDataResponse := &uploadResponse{}
	if err := json.NewDecoder(res.Body).Decode(metaDataResponse); err != nil {
		return errors.Wrap(err, "unmarshal response")
	}
	t := time.Now()
	c.metaData.Updated = &t
	c.metaData.Cluster = &metaDataResponse.Cluster
	c.metaData.MetaData = &metaDataResponse.MetaData

	log.Info().Str("fqdn", c.inventory.Cluster.Name).Int("status", res.StatusCode).Msg("uploaded inventory")

	return nil
}

func (c *InventoryCollection) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	err := json.NewEncoder(w).Encode(c.inventory)
	if err != nil {
		http.Error(w, errHTTPInternalError.JSON(), http.StatusInternalServerError)
		return
	}
}

func (c *InventoryCollection) ServeHTTPMeta(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	err := json.NewEncoder(w).Encode(c.metaData)
	if err != nil {
		http.Error(w, errHTTPInternalError.JSON(), http.StatusInternalServerError)
		return
	}
}

func (c *InventoryCollection) handleError(err error) {
	if err != nil {
		c.inventory.CollectionSucceeded = false
		c.inventory.CollectionErrors = append(c.inventory.CollectionErrors, err.Error())
		log.Error().Stack().Err(err).Msg("")
	}
}

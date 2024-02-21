package collect

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
const DefaultCollectionInterval = "1h"

type InventoryCollection struct {
	Inventory          *inventory.Inventory
	CollectionInterval string
	UploadInventory    bool
	Impersonate        string
	ServerAPIEndpoint  string
	TLSCrt             string
	TLSKey             string
	AuthEnabled        bool
	Signer             jose.Signer
}

func NewInventoryCollection(cfg config.Config) *InventoryCollection {
	i := &InventoryCollection{
		CollectionInterval: cfg.CollectionInterval,
		UploadInventory:    cfg.UploadInventory,
		Impersonate:        cfg.Impersonate,
		ServerAPIEndpoint:  fmt.Sprintf("%s/api/v1/inventory", cfg.ServerAPIEndpoint),
		TLSCrt:             cfg.TLSCrt,
		TLSKey:             cfg.TLSKey,
		AuthEnabled:        cfg.AuthEnabled,
	}
	if !i.AuthEnabled {
		log.Info().Msg("Authentication disabled")
		return i
	}
	if i.TLSCrt == "" {
		log.Info().Msg("No TLSCrt set. Authentication disabled")
		i.AuthEnabled = false
		return i
	}
	if i.TLSKey == "" {
		log.Error().Msg("No TLSKey set. Authentication disbaled.")
		i.AuthEnabled = false
		return i
	}
	if _, err := os.Stat(filepath.Clean(i.TLSCrt)); errors.Is(err, os.ErrNotExist) {
		log.Error().Msgf("TLSCrt file '%s' not found. Authentication disbaled.", i.TLSCrt)
		i.AuthEnabled = false
		return i
	}
	if _, err := os.Stat(filepath.Clean(i.TLSKey)); errors.Is(err, os.ErrNotExist) {
		log.Error().Msgf("TLSKey file '%s' not found. Authentication disbaled.", i.TLSKey)
		i.AuthEnabled = false
		return i
	}
	log.Info().Msg("Authentication enabled")
	i.AuthEnabled = true
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

	certificates, key, err := readKeyAndCertificates(c.TLSCrt, c.TLSKey)
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
	c.Signer = signer

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
		log.Info().Msgf("next iteration in %v at %v", r, t.Local().Format(time.DateTime))
		time.Sleep(r)
	}

	log.Info().Msg("entering inventory collection loop")
	for {
		c.Inventory = inventory.NewInventory()
		c.Inventory.CollectionSucceeded = true
		c.Inventory.ClientVersion = version.VERSION
		c.Inventory.ClientCommit = version.COMMIT
		cs, err := kubernetes.CreateK8SClient(c.Impersonate)
		if err != nil {
			log.Error().Err(err).Msg("creating clientset")
			c.Inventory.CollectionSucceeded = false
			c.Inventory.CollectionErrors = append(c.Inventory.CollectionErrors, err.Error())
			sleepNext()
			continue
		}

		log.Debug().Str("collect", "cluster").Msg("")
		c.handleError(CollectCluster(cs, c.Inventory))

		log.Debug().Str("collect", "scs").Msg("")
		c.handleError(CollectSCSMetadata(cs, c.Inventory))

		log.Debug().Str("collect", "namespace").Msg("")
		c.handleError(CollectNamespaces(cs, c.Inventory))

		log.Debug().Str("collect", "node").Msg("")
		c.handleError(CollectNodes(cs, c.Inventory))

		log.Debug().Str("collect", "storage").Msg("")
		c.handleError(CollectStorage(cs, c.Inventory))

		log.Debug().Str("collect", "components").Msg("")
		c.handleError(CollectCustomResources(cs, c.Inventory))

		log.Debug().Str("collect", "workload").Msg("")
		c.handleError(CollectWorkloads(cs, c.Inventory))

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

	var (
		payload     []byte
		err         error
		contentType string = "application/json; charset=UTF-8"
	)

	payload, err = json.Marshal(c.Inventory)
	if err != nil {
		return errors.Wrap(err, "marshaling inventory")
	}

	if c.AuthEnabled {
		jwt, err := c.Signer.Sign(payload)
		if err != nil {
			return errors.Wrap(err, "signing inventory")
		}
		payload = []byte(jwt.FullSerialize())
		contentType = "application/jose+json"
	}

	req, err := http.NewRequest("PUT", c.ServerAPIEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", contentType)
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

func (c *InventoryCollection) handleError(err error) {
	if err != nil {
		c.Inventory.CollectionSucceeded = false
		c.Inventory.CollectionErrors = append(c.Inventory.CollectionErrors, err.Error())
		log.Error().Stack().Err(err).Msg("")
	}
}

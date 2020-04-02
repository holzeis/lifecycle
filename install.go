package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// ChaincodeServerUserData represents the connection json structure.
type ChaincodeServerUserData struct {
	Address            string `json:"address"`
	DialTimeout        string `json:"dial_timeout"`
	TLSRequired        bool   `json:"tls_required"`
	ClientAuthRequired bool   `json:"client_auth_required"`
	ClientKey          string `json:"client_key"`  // PEM encoded client key
	ClientCert         string `json:"client_cert"` // PEM encoded client certificate
	RootCert           string `json:"root_cert"`   // PEM encoded peer chaincode certificate
}

// PackageMetadata represents the metadata json structure.
type PackageMetadata struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

func (l *Lifecycle) createMetadata(path string) error {
	metadataAsBytes, err := json.Marshal(PackageMetadata{
		Label: l.Chaincode,
		Path:  "",
		Type:  "external",
	})
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%v/metadata.json", path), metadataAsBytes, 0644); err != nil {
		return err
	}
	return nil
}

func (l *Lifecycle) createConnection(path string) error {
	connectionAsBytes, err := json.Marshal(ChaincodeServerUserData{
		Address:     fmt.Sprintf("%v:7052", l.Chaincode),
		DialTimeout: "10s",
		TLSRequired: false,
	})
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%v/connection.json", path), connectionAsBytes, 0644); err != nil {
		return err
	}
	return nil
}

// Install installs the chaincode to the network using the nodes discovered by the discovery service. Performs a http request for each msp which is not the current.
func (l *Lifecycle) Install() error {
	for _, node := range l.Nodes {
		if node.MSPID == l.MSPID {
			// if msp is local msp, no need to make an http request
			if err := l.install(); err != nil {
				return err
			}
		} else {
			// ask participants to approve the chaincode installation
			resp, err := http.Get(fmt.Sprintf("%v://lifecycle.%v:%v/install/%v", "http", node.Host, "8090", l.Chaincode))
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("%v returned status code %v", node.MSPID, resp.StatusCode)
			}
		}
		logger.Infof("%v successfully installed the chaincode", node.MSPID)
	}

	return nil
}

func (l *Lifecycle) install() error {
	txid := uuid.New()
	path, err := ioutil.TempDir("", txid.String())
	if err != nil {
		return err
	}

	if err := l.createConnection(path); err != nil {
		return err
	}

	if err := l.createMetadata(path); err != nil {
		return err
	}

	if _, err := l.execute(fmt.Sprintf("tar -czf %[1]v/code.tar.gz -C %[1]v connection.json", path)); err != nil {
		return err
	}

	if _, err := l.execute(fmt.Sprintf("tar -czf %[1]v/%[2]v.tgz -C %[1]v code.tar.gz metadata.json", path, l.Chaincode)); err != nil {
		return err
	}

	command := []string{
		"peer lifecycle chaincode install",
		fmt.Sprintf("%v/%v.tgz", path, l.Chaincode),
		fmt.Sprintf("--peerAddresses %v", os.Getenv("CORE_PEER_ADDRESS")),
		fmt.Sprintf("--tlsRootCertFiles %v", os.Getenv("CORE_PEER_TLS_ROOTCERT_FILE")),
	}

	response, err := l.execute(strings.Join(command, " "))
	if err != nil {
		return err
	}

	l.CCID, _ = response.findInLogs(fmt.Sprintf("%v:\\w*", l.Chaincode))
	return nil
}

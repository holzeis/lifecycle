package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Node ...
type Node struct {
	MSPID  string
	Host   string
	RootCA string
}

// NewNode builds a new node.
func NewNode(mspID, endpoint string) Node {
	return Node{
		MSPID: mspID,
		// the peer endpoint is added like peer-0.peer.host. we need to split the prefixes in order to get the host.
		Host: strings.Join(strings.Split(strings.Split(endpoint, ":")[0], ".")[2:], "."),
	}
}

// Discover discovers the nodes within the network.
func (l *Lifecycle) Discover() (err error) {
	keystore, err := findKeystore()
	if err != nil {
		return err
	}

	signcert, err := findSigncert()
	if err != nil {
		return err
	}

	peers, err := l.peers(keystore, signcert)
	if err != nil {
		return err
	}

	config, err := l.config(keystore, signcert)
	if err != nil {
		return err
	}

	for _, peer := range peers {
		// build nodes from peers and config
		node := NewNode(peer["MSPID"].(string), peer["Endpoint"].(string))

		msp := config["msps"].(map[string]interface{})[node.MSPID].(map[string]interface{})
		rootCA, err := base64.StdEncoding.DecodeString(msp["tls_root_certs"].([]interface{})[0].(string))
		if err != nil {
			logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		}
		rootCAFile, err := ioutil.TempFile("", fmt.Sprintf("%v", node.MSPID))
		if err != nil {
			return err
		}
		// saving root ca to file for cli processing.
		if _, err = rootCAFile.Write(rootCA); err != nil {
			return err
		}

		// root ca has to be provided as file to the cli, hence we save the path to the saved root ca to the node.
		node.RootCA = rootCAFile.Name()
		l.Nodes = append(l.Nodes, node)
	}

	return nil
}

func (l *Lifecycle) peers(keystore, signcert string) (peers []map[string]interface{}, err error) {
	command := []string{
		"discover peers",
		fmt.Sprintf("--channel %v", l.Channel),
		fmt.Sprintf("--server %v", os.Getenv("CORE_PEER_ADDRESS")),
		fmt.Sprintf("--peerTLSCA %v", os.Getenv("CORE_PEER_TLS_ROOTCERT_FILE")),
		fmt.Sprintf("--userKey %v", filepath.Join(os.Getenv("CORE_PEER_MSPCONFIGPATH"), "keystore", keystore)),
		fmt.Sprintf("--userCert %v", filepath.Join(os.Getenv("CORE_PEER_MSPCONFIGPATH"), "signcerts", signcert)),
		fmt.Sprintf("--MSP %v", l.MSPID),
	}

	response, err := l.execute(strings.Join(command, " "))
	if err != nil {
		return peers, err
	}

	err = json.Unmarshal(response.Output.Bytes(), &peers)
	return peers, err
}

func (l *Lifecycle) config(keystore, signcert string) (config map[string]interface{}, err error) {
	command := []string{
		"discover config",
		fmt.Sprintf("--channel %v", l.Channel),
		fmt.Sprintf("--server %v", os.Getenv("CORE_PEER_ADDRESS")),
		fmt.Sprintf("--peerTLSCA %v", os.Getenv("CORE_PEER_TLS_ROOTCERT_FILE")),
		fmt.Sprintf("--userKey %v", filepath.Join(os.Getenv("CORE_PEER_MSPCONFIGPATH"), "keystore", keystore)),
		fmt.Sprintf("--userCert %v", filepath.Join(os.Getenv("CORE_PEER_MSPCONFIGPATH"), "signcerts", signcert)),
		fmt.Sprintf("--MSP %v", l.MSPID),
	}

	response, err := l.execute(strings.Join(command, " "))
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(response.Output.Bytes(), &config)
	return config, err
}

func findKeystore() (string, error) {
	// read keystore file as the name is generated
	files, err := ioutil.ReadDir(filepath.Join(os.Getenv("CORE_PEER_MSPCONFIGPATH"), "keystore"))
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("Missing keystore")
	}
	return files[0].Name(), nil
}

func findSigncert() (string, error) {
	// read signcerts file as the name is generated
	files, err := ioutil.ReadDir(filepath.Join(os.Getenv("CORE_PEER_MSPCONFIGPATH"), "signcerts"))
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("Missing signcert")
	}
	return files[0].Name(), nil
}

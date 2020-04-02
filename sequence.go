package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// QueryCommitted ...
type QueryCommitted struct {
	Sequence int    `json:"sequence"`
	Version  string `json:"version"`
}

// NextSequence ...
func (l *Lifecycle) NextSequence() error {
	command := []string{
		"peer lifecycle chaincode querycommitted",
		fmt.Sprintf("--channelID %v", l.Channel),
		fmt.Sprintf("--name %v", l.Chaincode),
		fmt.Sprintf("-o %v", os.Getenv("ORDERER_ADDRESS")),
		fmt.Sprintf("--tls %v", "true"),
		fmt.Sprintf("--cafile %v", os.Getenv("ORDERER_CA")),
		fmt.Sprintf("--peerAddresses %v", os.Getenv("CORE_PEER_ADDRESS")),
		fmt.Sprintf("--tlsRootCertFiles %v", os.Getenv("CORE_PEER_TLS_ROOTCERT_FILE")),
		"-O json",
	}

	response, err := l.execute(strings.Join(command, " "))
	if err == nil {
		var committed QueryCommitted
		err = json.Unmarshal(response.Output.Bytes(), &committed)
		l.Sequence = committed.Sequence + 1
		return err
	}
	return nil
}

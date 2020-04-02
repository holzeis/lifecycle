package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GetCCID ...
func (l *Lifecycle) GetCCID() (err error) {
	command := []string{
		"peer lifecycle chaincode queryinstalled",
		fmt.Sprintf("--peerAddresses %v", os.Getenv("CORE_PEER_ADDRESS")),
		fmt.Sprintf("--tlsRootCertFiles %v", os.Getenv("CORE_PEER_TLS_ROOTCERT_FILE")),
		"-O json",
	}

	response, err := l.execute(strings.Join(command, " "))
	if err != nil {
		return err
	}

	var chaincodes map[string]interface{}
	if err = json.Unmarshal(response.Output.Bytes(), &chaincodes); err != nil {
		return err
	}

	installed := chaincodes["installed_chaincodes"].([]interface{})
	for _, inst := range installed {
		mapped := inst.(map[string]interface{})
		if chaincode := mapped["label"]; chaincode != l.Chaincode {
			// skip if the installed chaincode does not equal the requested chaincode
			continue
		}

		if _, ok := mapped["references"]; !ok {
			// skip if no channel is referenced
			continue
		}

		if _, ok := mapped["references"].(map[string]interface{})[l.Channel]; !ok {
			// skip if the requested channel is not referenced.
			continue
		}

		l.CCID = mapped["package_id"].(string)

		return nil
	}

	return nil
}

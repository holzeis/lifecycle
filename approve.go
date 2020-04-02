package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Approve ...
func (l *Lifecycle) Approve() error {
	for _, node := range l.Nodes {
		if node.MSPID == l.MSPID {
			// if msp is local msp, no need to make an http request
			if err := l.approve(); err != nil {
				return err
			}
		} else {
			// ask participants to approve the chaincode installation
			resp, err := http.Get(fmt.Sprintf("%v://lifecycle.%v:%v/%v/approve/%v/%v/%v", "http", node.Host, "8090", l.Channel, l.Chaincode, l.Sequence, l.CCID))
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("%v returned status code %v", node.MSPID, resp.StatusCode)
			}
		}
		logger.Infof("%v approved the chaincode installation", node.MSPID)
	}

	return nil
}

func (l *Lifecycle) approve() error {
	if l.checkIfChaincodeIsApproved() {
		logger.Warnf("%v with sequence %v has already been approved on %v by %v", l.Chaincode, l.Sequence, l.Channel, l.MSPID)
		return nil
	}

	command := []string{
		"peer lifecycle chaincode approveformyorg",
		fmt.Sprintf("--channelID %v", l.Channel),
		fmt.Sprintf("--name %v", l.Chaincode),
		fmt.Sprintf("--version %v", "1.0"),
		fmt.Sprintf("--package-id %v", l.CCID),
		fmt.Sprintf("--sequence %v", l.Sequence),
		fmt.Sprintf("-o %v", os.Getenv("ORDERER_ADDRESS")),
		fmt.Sprintf("--tls %v", "true"),
		fmt.Sprintf("--cafile %v", os.Getenv("ORDERER_CA")),
	}
	// "--init-required"

	// approve chaincode installation
	_, err := l.execute(strings.Join(command, " "))
	return err
}

func (l *Lifecycle) checkIfChaincodeIsApproved() bool {
	command := []string{
		"peer lifecycle chaincode checkcommitreadiness",
		fmt.Sprintf("--channelID %v", l.Channel),
		fmt.Sprintf("--name %v", l.Chaincode),
		fmt.Sprintf("--version %v", "1.0"),
		fmt.Sprintf("--sequence %v", l.Sequence),
		fmt.Sprintf("-o %v", os.Getenv("ORDERER_ADDRESS")),
		fmt.Sprintf("--tls %v", "true"),
		fmt.Sprintf("--cafile %v", os.Getenv("ORDERER_CA")),
		"-O json",
	}
	// "--init-required"

	// approve chaincode installation
	response, err := l.execute(strings.Join(command, " "))
	if err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		return false
	}

	var approvals map[string]interface{}
	if err := json.Unmarshal(response.Output.Bytes(), &approvals); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		return false
	}

	return approvals["approvals"].(map[string]interface{})[l.MSPID].(bool)
}

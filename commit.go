package main

import (
	"fmt"
	"os"
	"strings"
)

// Commit commits the chaincode to the network, using the nodes discovered by the discovery services.
func (l *Lifecycle) Commit() error {
	command := []string{
		"peer lifecycle chaincode commit",
		fmt.Sprintf("--channelID %v", l.Channel),
		fmt.Sprintf("--name %v", l.Chaincode),
		fmt.Sprintf("--version %v", "1.0"),
		fmt.Sprintf("--sequence %v", l.Sequence),
		fmt.Sprintf("-o %v", os.Getenv("ORDERER_ADDRESS")),
		fmt.Sprintf("--tls %v", "true"),
		fmt.Sprintf("--cafile %v", os.Getenv("ORDERER_CA")),
	}
	// --init-required

	for _, node := range l.Nodes {
		if node.Name != "peer-0" {
			continue
		}
		fmt.Println(fmt.Sprintf("--peerAddresses peer.%v:%v", node.Host, "7051"))
		command = append(command, fmt.Sprintf("--peerAddresses peer.%v:%v", node.Host, "7051"))
		command = append(command, fmt.Sprintf("--tlsRootCertFiles %v", node.RootCA))
	}

	// committing chaincode installation
	_, err := l.execute(strings.Join(command, " "))
	return err
}

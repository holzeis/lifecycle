# Chaincode lifecycle

Wraps the peer [chaincode lifecycle](https://hyperledger-fabric.readthedocs.io/en/release-2.0/commands/peerlifecycle.html) cli api into an http server.

* Supports chaincode installations as an external service.
* Coordinates chaincode installations within the business network.
* Provides an api to query the currently installed / committed ccid.

## API

The following lists the accessible endpoints exposed by the lifecycle service.

### POST /{channel}/deploy/{chaincode}

Deploys a chaincode to the network using the discovery service to find nodes participating in the channel. A connection and metadata json is created based on the given parameters. *Please note, that the chaincode as external service is expected to be accessible on {chaincode}:7052.*

### GET /install/{chaincode}

Installs a chaincode as external service to the given peer

### GET /{channel}/approve/{chaincode}/{sequence}/{ccid}

Approves a chaincode installation for the given channel, chaincode, sequence number and ccid (package id).

### GET /{channel}/installed/{chaincode}

Returnes the installed and committed chaincode on the given channel. Returns 404 if the chaincode has not been installed on that channel.

## Used environment variables

The following environment variables need to be set in order for the lifecycle service to work properly.

|Environment Variable|Description|
|--------------------|-----------|
|FABRIC_LOGGING_SPEC|sets the log level e.g. INFO|
|CORE_PEER_LOCALMSPID|the msp of the organization|
|ORDERER_ADDRESS|the address of the orderer|
|ORDERER_CA|the ca of the orderer for tls communication|
|CORE_PEER_ADDRESS|the address of the peer|
|CORE_PEER_TLS_ENABLED|whether or not tls should be used (needs to be true)|
|CORE_PEER_MSPCONFIGPATH|the path to the users msp config|
|CORE_PEER_TLS_CERT_FILE|the path to the peers cert file|
|CORE_PEER_TLS_ROOTCERT_FILE|the path to the peers root cert file|

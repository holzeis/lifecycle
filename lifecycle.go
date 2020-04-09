package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("lifecycle")

// Lifecycle keeping all data required for the lifecycle cli commands
type Lifecycle struct {
	MSPID string

	Channel   string
	Chaincode string
	Sequence  int
	CCID      string
	Nodes     []Node
}

// NewLifecycle builds a new lifecycle struct
func NewLifecycle(vars map[string]string) Lifecycle {
	sequence := 1
	if seq, ok := vars["sequence"]; ok {
		sequence, _ = strconv.Atoi(seq)
	}

	return Lifecycle{
		Chaincode: vars["chaincode"],
		Channel:   vars["channel"],
		MSPID:     os.Getenv("CORE_PEER_LOCALMSPID"),
		Sequence:  sequence,
		CCID:      vars["ccid"],
	}
}

// Deploy deploys a chaincode as external service to the network.
func Deploy(w http.ResponseWriter, req *http.Request) {
	lifecycle := NewLifecycle(mux.Vars(req))

	logger.Info("Discovering the network")
	if err := lifecycle.Discover(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	logger.Infof("Found %v nodes", len(lifecycle.Nodes))

	logger.Infof("Installing %v", lifecycle.Chaincode)
	if err := lifecycle.Install(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	logger.Infof("Installed chaincode with ccid: %v", lifecycle.CCID)

	logger.Info("Calculating next sequence number")
	if err := lifecycle.NextSequence(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	logger.Infof("Next sequence number: %v", lifecycle.Sequence)

	logger.Infof("Approving %v on %v", lifecycle.CCID, lifecycle.Channel)
	if err := lifecycle.Approve(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	logger.Infof("Successfully approved %v with ccid %v on %v", lifecycle.Chaincode, lifecycle.CCID, lifecycle.Channel)

	logger.Infof("Committing %v to %v", lifecycle.CCID, lifecycle.Channel)
	if err := lifecycle.Commit(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
	}
}

// Install installs a chaincode as external service to the given peer.
func Install(w http.ResponseWriter, req *http.Request) {
	lifecycle := NewLifecycle(mux.Vars(req))
	if err := lifecycle.install(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
	}
	logger.Infof("Successfully installed %v with ccid %v", lifecycle.Chaincode, lifecycle.CCID)
}

// Approve approves the given chaincode for the given channel and ccid.
func Approve(w http.ResponseWriter, req *http.Request) {
	lifecycle := NewLifecycle(mux.Vars(req))

	if err := lifecycle.approve(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
	}
	logger.Infof("Successfully approved %v with ccid %v[%v] on %v", lifecycle.Chaincode, lifecycle.CCID, lifecycle.Sequence, lifecycle.Channel)
}

// Installed returns the ccid of the requested chaincode and channel.
func Installed(w http.ResponseWriter, req *http.Request) {
	lifecycle := NewLifecycle(mux.Vars(req))

	if err := lifecycle.GetCCID(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err.Error()))
		http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
	}
	if lifecycle.CCID == "" {
		logger.Warnf("CCID for %v could not be found on %v", lifecycle.Chaincode, lifecycle.Channel)
		http.Error(w, fmt.Sprintf("CCID for %v could not be found on %v", lifecycle.Chaincode, lifecycle.Channel), http.StatusNotFound)
		return
	}

	logger.Infof("Found package id: %v on %v for %v", lifecycle.CCID, lifecycle.Channel, lifecycle.Chaincode)
	fmt.Fprintf(w, "%v", lifecycle.CCID)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{channel}/deploy/{chaincode}", Deploy).Methods("GET")
	r.HandleFunc("/install/{chaincode}", Install).Methods("GET")
	r.HandleFunc("/{channel}/approve/{chaincode}/{sequence}/{ccid}", Approve).Methods("GET")
	r.HandleFunc("/{channel}/installed/{chaincode}", Installed).Methods("GET")
	server := &http.Server{Addr: ":8090", Handler: r}

	go func() {
		logger.Info("Listening on 0.0.0.0:8090")
		logger.Fatal(server.ListenAndServe())
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop
	logger.Warn("Stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal(err)
	}
}

package intranet

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/getblank/wango"
)

var (
	srAddress string
	tqAddress string
	httpPort  string
	srClient  *wango.Wango
	srLocker  sync.RWMutex
)

// Init is the main entry point for the intranet package
func Init(addr, port string) {
	srAddress = addr
	httpPort = port
	go connectToSr()
	startHTTPServer()
}

type service struct {
	Type    string `json:"type"`
	Address string `json:"address"`
	Port    string `json:"port"`
}

func connectedToSR(w *wango.Wango) {
	log.Info("Connected to SR: ", srAddress)
	srLocker.Lock()
	srClient = w
	srLocker.Unlock()

	srClient.Call("register", map[string]interface{}{"type": "fileStore", "port": httpPort})
}

func connectToSr() {
	reconnectChan := make(chan struct{})
	for {
		log.Info("Attempt to connect to SR: ", srAddress)
		client, err := wango.Connect(srAddress, "http://getblank.net")
		if err != nil {
			log.Warn("Can'c connect to service registry: " + err.Error())
			time.Sleep(time.Second)
			continue
		}

		client.SetSessionCloseCallback(func(c *wango.Conn) {
			log.Warn("Disconnected from service registry")
			srLocker.Lock()
			srClient = nil
			srLocker.Unlock()
			reconnectChan <- struct{}{}
		})
		connectedToSR(client)
		<-reconnectChan
	}
}

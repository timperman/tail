package plugin

import (
	"encoding/json"
	"github.com/timperman/tail/driver"
	"github.com/timperman/tail/stream"
	"log"
	"net/http"
)

type Handshake struct {
	Implements []string
}

func Start(addr string, root string) {
	broker := stream.NewServer()

	http.Handle("/stream", broker)
	http.HandleFunc("/Plugin.Activate", activate)

	v, err := driver.New(root, broker.Notifier)
  if err != nil {
    log.Fatal("error creating driver: %v\n", err)
  }

	m := map[string]map[string]func(http.ResponseWriter, *http.Request){
		"POST": {
			"/VolumeDriver.Create":  v.Create,
			"/VolumeDriver.Remove":  v.Remove,
			"/VolumeDriver.Mount":   v.Mount,
			"/VolumeDriver.Unmount": v.Unmount,
			"/VolumeDriver.Path":    v.Path,
		},
	}

	for method, routes := range m {
		for route, f := range routes {
			http.HandleFunc(route, handleFuncByMethod(method, f))
		}
	}

	log.Fatal(http.ListenAndServe(addr, nil))
}

func activate(w http.ResponseWriter, r *http.Request) {
	log.Println("Activate call")
	if b, err := json.Marshal(&Handshake{Implements: []string{"VolumeDriver"}}); err == nil {
		w.Write(b)
	}
}

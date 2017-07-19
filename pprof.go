package main

import (
	"net/http"
)

func setupPProf(port string) {
	addr := "0.0.0.0:" + port
	go func() {
		http.ListenAndServe(addr, nil)
	}()
}

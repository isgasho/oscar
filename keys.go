package main

import (
	"net/http"

	"pijun.io/oscar/sodium"
)

var oscarKeyPair sodium.KeyPair
var oscarSymKey []byte

func getServerPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	sendSuccess(w, struct {
		Key encodableBytes `json:"public_key"`
	}{Key: oscarKeyPair.Public})
}

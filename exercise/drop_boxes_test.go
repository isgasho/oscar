package main

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"zood.xyz/oscar/sodium"
)

const dropBoxIDSize = 16
const watchCommand = 1

func TestPackageDropping(t *testing.T) {
	user := createUserOnServer(t)
	accessToken := login(user, t)

	pkg := []byte("N. Bluth")
	boxID := make([]byte, dropBoxIDSize)
	sodium.Random(boxID)

	req, _ := http.NewRequest(http.MethodPut, apiRoot+"/alpha/drop-boxes/"+hex.EncodeToString(boxID), bytes.NewReader(pkg))
	req.Header.Add("X-Oscar-Access-Token", accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Incorrect status code: %d", resp.StatusCode)
	}

	// now try to pick it up
	req, _ = http.NewRequest(http.MethodGet, apiRoot+"/alpha/drop-boxes/"+hex.EncodeToString(boxID), nil)
	req.Header.Add("X-Oscar-Access-Token", accessToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Incorrect status code: %d", resp.StatusCode)
	}

	rcvdPkg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rcvdPkg, pkg) {
		t.Fatal("Downloaded package doesn't match what was dropped off")
	}
}

func TestPackageWatching(t *testing.T) {
	boxID := make([]byte, dropBoxIDSize)
	sodium.Random(boxID)
	msgChan := make(chan []byte)

	conn, _, err := websocket.DefaultDialer.Dial("ws://"+apiAddress+"/alpha/drop-boxes/watch", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Reader
	go func() {
		for {
			msgType, buf, err := conn.ReadMessage()
			if err != nil {
				// probably becaused the connection was closed
				return
			}
			if msgType != websocket.BinaryMessage {
				t.Fatal("Invalid message type")
				return
			}
			if len(buf) == 0 {
				t.Fatal("Message is empty")
				return
			}

			// sanity check on the message
			if len(buf) < 1+dropBoxIDSize+1 {
				t.Fatalf("Message was too short - only %d bytes", len(buf))
				return
			}

			// make sure it's a watch response
			if buf[0] != watchCommand {
				t.Fatalf("Unknown message/command: %d", buf[0])
			}
			rcvdBoxID := buf[1 : dropBoxIDSize+1]
			if !bytes.Equal(rcvdBoxID, boxID) {
				t.Fatalf("Box ids didn't match: %s != %s",
					hex.EncodeToString(rcvdBoxID),
					hex.EncodeToString(boxID))
			}
			rcvdPkg := buf[1+dropBoxIDSize:]
			msgChan <- rcvdPkg
			conn.Close()
			close(msgChan)
			return
		}
	}()
	// write the watch command
	watchCmd := append([]byte{1}, boxID...)
	if err = conn.WriteMessage(websocket.BinaryMessage, watchCmd); err != nil {
		t.Fatal(err)
	}

	user := createUserOnServer(t)
	accessToken := login(user, t)

	// drop a package
	pkg := []byte("N. Bluth")
	req, _ := http.NewRequest(http.MethodPut, apiRoot+"/alpha/drop-boxes/"+hex.EncodeToString(boxID), bytes.NewReader(pkg))
	req.Header.Add("X-Oscar-Access-Token", accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Incorrect status code: %d", resp.StatusCode)
	}

	select {
	case rcvdPkg := <-msgChan:
		if !bytes.Equal(rcvdPkg, pkg) {
			t.Fatalf("package didn't match: %s != %s",
				string(rcvdPkg),
				string(pkg))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Didn't receive the package in time")
	}
	conn.Close()
}
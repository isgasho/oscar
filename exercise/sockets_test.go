package main

import (
	"bytes"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"zood.xyz/oscar/sodium"
)

const (
	socketCmdNop    byte = 0
	socketCmdWatch       = 1
	socketCmdIgnore      = 2
)

type socketClient struct {
	conn   *websocket.Conn
	inbox  <-chan []byte
	outbox chan<- []byte
}

func (sc *socketClient) readConn(writableInbox chan []byte, t *testing.T) {
	msgType, buf, err := sc.conn.ReadMessage()
	if err != nil {
		return
	}
	if msgType != websocket.BinaryMessage {
		t.Fatal("socket received a non-binary message")
	}
	writableInbox <- buf
}

func (sc *socketClient) start(t *testing.T) {
	writableInbox := make(chan []byte, 5)
	sc.inbox = writableInbox
	go sc.readConn(writableInbox, t)

	readableOutbox := make(chan []byte, 5)
	sc.outbox = readableOutbox
	go sc.writeConn(readableOutbox, t)
}

func (sc *socketClient) writeConn(readableOutbox chan []byte, t *testing.T) {
	for msg := range readableOutbox {
		if msg == nil {
			return
		}
		err := sc.conn.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSocketServer(t *testing.T) {
	user := createUserOnServer(t)
	token := login(user, t)

	hdr := make(http.Header)
	hdr.Add("Sec-Websocket-Protocol", token)
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+apiAddress+"/alpha/sockets", hdr)
	if err != nil {
		t.Fatal(err)
	}
	sc := socketClient{conn: conn}
	sc.start(t)

	// make up a drop box address
	boxID := make([]byte, dropBoxIDSize)
	sodium.Random(boxID)

	pkg1 := []byte("Hello, my world!")
	dropPackage(pkg1, boxID, token, t)

	// send a watch command for that box
	sc.outbox <- append([]byte{socketCmdWatch}, boxID...)

	// we should get a box notification message back on the socket
	select {
	case <-time.After(200 * time.Millisecond):
		t.Fatal("did not receive box notification back in time")
	case rcvdPkg := <-sc.inbox:
		shouldBe := append([]byte{socketCmdWatch}, boxID...)
		shouldBe = append(shouldBe, pkg1...)
		if !bytes.Equal(rcvdPkg, shouldBe) {
			t.Fatal("Received package did not match expected")
		} else {
			log.Printf("the package matched")
		}
	}

	// create another user to send us a message, so we receive it over the socket
	otherUser := createUserOnServer(t)
	otherToken := login(otherUser, t)
	msg := outboundMessage{
		CipherText: []byte("some cipher text"),
		Nonce:      []byte("some nonce"),
		Urgent:     true,
		Transient:  true,
	}
	sendMessage(msg, user.publicID, otherToken, t)

	// Our next notification on the socket should be a message from the user.
	select {
	case <-time.After(200 * time.Millisecond):
		t.Fatal("did not receive the message notification in time")
	case rcvdMsg := <-sc.inbox:

	}
}
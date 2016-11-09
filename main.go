package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// Debug contains whether the server is running in debug mode
var Debug = false

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	port := flag.Int("port", 80, "Listening port for server")
	debug := flag.Bool("debug", false, "Enables additional log output")
	sqlDSN := flag.String(
		"sqldsn",
		"",
		"DSN to SQL server e.g. username:password@protocol(address)/dbname?param=value")
	kvDBPath := flag.String("kvdb", "", "Path to key-value database file")
	flag.Parse()
	Debug = *debug

	gFCMServerKey = os.Getenv("FCM_SERVER_KEY")
	if gFCMServerKey == "" {
		log.Fatal("$FCM_SERVER_KEY is missing/empty")
	}

	err := initDB(*sqlDSN)
	if err != nil {
		log.Fatalf("Error initializing SQL db: %v", err)
	}

	err = initKVDB(*kvDBPath)
	if err != nil {
		log.Fatalf("Error initializing key-value db: %v", err)
	}

	r := mux.NewRouter()
	alphaRouter := r.PathPrefix("/alpha").Subrouter()
	installEndPoints(alphaRouter)

	// playground()

	hostAddress := fmt.Sprintf(":%d", *port)
	server := http.Server{
		Addr:         hostAddress,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Printf("Starting server on port %d", *port)
	server.ListenAndServe()
}

func installEndPoints(r *mux.Router) {
	r.Handle("/users", newRESTFunc(searchUsersHandler)).Methods("GET")
	r.Handle("/users", newRESTFunc(createUserHandler)).Methods("POST")
	r.Handle("/users/me/fcm-tokens", newRESTFunc(addFCMTokenHandler)).Methods("POST")
	r.Handle("/users/me/fcm-tokens/{token}", newRESTFunc(deleteFCMTokenHandler)).Methods("DELETE")
	r.Handle("/users/{public_id}", newRESTFunc(getUserInfoHandler)).Methods("GET")
	r.Handle("/users/{public_id}/messages", newRESTFunc(sendMessageToUserHandler)).Methods("POST")
	r.Handle("/users/{public_id}/public-key", newRESTFunc(getUserPublicKeyHandler)).Methods("GET")

	r.Handle("/messages", newRESTFunc(getMessagesHandler)).Methods("GET")
	r.Handle("/messages/{msg_id}/processed", newRESTFunc(editMessageHandler)).Methods("PUT")

	// this has to come first, so it has a chance to match before the box_id urls
	r.Handle("/drop-boxes/watch", newRESTFunc(createPackageWatcherHandler)).Methods("GET")
	r.Handle("/drop-boxes/{box_id}", newRESTFunc(pickUpPackageHandler)).Methods("GET")
	r.Handle("/drop-boxes/{box_id}", newRESTFunc(dropPackageHandler)).Methods("PUT")

	r.Handle("/sessions/{username}/challenge", newRESTFunc(createAuthChallengeHandler)).Methods("POST")
	r.Handle("/sessions/{username}/challenge-response", newRESTFunc(authChallengeResponseHandler)).Methods("POST")

	r.Handle("/goroutine-stacks", newRESTFunc(goroutineStacksHandler)).Methods("GET")
	r.Handle("/test", newRESTFunc(testHandler)).Methods("GET")
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	pushMessageToUser(22, 16)
	// sendFirebaseMessage(16, nil)
}

func playground() {
	// rdr := bytes.NewReader([]byte{'w', 'o', 'r', 'd'})
	// var token string
	// err := json.NewDecoder(rdr).Decode(&token)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("decoded: %v", token)
}
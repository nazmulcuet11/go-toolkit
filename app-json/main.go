package main

import (
	"log"
	"net/http"

	"github.com/nazmulcuet11/go-toolkit/toolkit"
)

type RequestPayload struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

type ResponsePayload struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/receive-post", receivePost)
	mux.HandleFunc("/remote-service", remoteService)
	mux.HandleFunc("/simulate-service", simulateService)

	log.Println("starting server at port: 8081")
	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func receivePost(w http.ResponseWriter, r *http.Request) {
	requestPayload := RequestPayload{}
	t := toolkit.Tools{}
	err := t.ReadJSON(w, r, &requestPayload)
	if err != nil {
		t.ErrorJSON(w, err)
		return
	}

	responsePayload := ResponsePayload{
		Message: "hit the handler",
	}
	err = t.WriteJSON(w, http.StatusOK, responsePayload)
	if err != nil {
		log.Println(err)
	}
}

func remoteService(w http.ResponseWriter, r *http.Request) {
	requestPayload := RequestPayload{}
	t := toolkit.Tools{}
	err := t.ReadJSON(w, r, &requestPayload)
	if err != nil {
		t.ErrorJSON(w, err)
		return
	}

	_, statusCode, err := t.PushJSONToRemote("http://localhost:8081/simulate-service", requestPayload)
	if err != nil {
		t.ErrorJSON(w, err)
		return
	}

	responsePayload := ResponsePayload{
		Message:    "hit the handler",
		StatusCode: statusCode,
	}
	err = t.WriteJSON(w, http.StatusOK, responsePayload)
	if err != nil {
		log.Println(err)
	}
}

func simulateService(w http.ResponseWriter, r *http.Request) {
	payload := ResponsePayload{
		Message: "OK",
	}
	t := toolkit.Tools{}
	_ = t.WriteJSON(w, http.StatusOK, payload)
}

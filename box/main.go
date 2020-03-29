package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//Mail - Stores all the attributes for a given email
//Includes an ID primary key
type Mail struct {
	ID   string `json:"id"`
	To   string `json:"to"`
	From string `json:"from"`
	Body string `json:"body"`
}

//ClientError - Provides structure to user errors
type ClientError struct {
	StatusCode     int    `json:"statusCode"`
	Detail         string `json:"detail"`
	Recommendation string `json:"recommendation"`
}

//Retrieves the environment variables to distinguish multiple email servers
var serverIndex string = os.Getenv("SERVER_INDEX")
var domainName string = os.Getenv("DOMAIN_NAME")

//Initialize mails var as a slice Mail struct
var postbox []Mail

//Get all mail for a given user
func findAllByUser(w http.ResponseWriter, r *http.Request) {
	var subbox []Mail
	params := mux.Vars(r)

	for _, mail := range postbox {
		if mail.From == fmt.Sprintf("%s@%s", params["user"], domainName) || mail.To == fmt.Sprintf("%s@%s", params["user"], domainName) {
			subbox = append(subbox, mail)
		}
	}

	json.NewEncoder(w).Encode(subbox)
}

//Retrieves all mail
func findAll(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(postbox)
}

// Gets a single mail based on ID
func find(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	for _, mail := range postbox {
		if mail.ID == params["id"] {
			json.NewEncoder(w).Encode(mail)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ClientError{http.StatusNotFound,
			"Not email with this ID found.",
			"Retrieve all mail to check existing IDs!"})
	}
}

//Create new mail
func save(w http.ResponseWriter, r *http.Request) {
	var mail Mail
	_ = json.NewDecoder(r.Body).Decode(&mail)
	postbox = append(postbox, mail)

	json.NewEncoder(w).Encode(mail)
}

//Deletes mail by ID
func delete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	for i, mail := range postbox {
		if mail.ID == params["id"] {
			postbox[i] = postbox[len(postbox)-1]
			postbox = postbox[:len(postbox)-1]

			json.NewEncoder(w).Encode(postbox)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ClientError{http.StatusNotFound,
			"Not email with this ID found.",
			"Retrieve all mail to check existing IDs!"})
	}
}

//Ensures all routes return application/json content
func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()
	r.Use(contentTypeMiddleware)

	r.HandleFunc("/mail/{user}", findAllByUser).Methods("GET")
	r.HandleFunc("/mail", findAll).Methods("GET")
	r.HandleFunc("/mail/{id}", find).Methods("GET")
	r.HandleFunc("/mail", save).Methods("POST")
	r.HandleFunc("/mail/{id}", delete).Methods("DELETE")

	fmt.Printf("Server %s: Box initialized!\n", serverIndex)
	log.Fatal(http.ListenAndServe(":8000", r))
}

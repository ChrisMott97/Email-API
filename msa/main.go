package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

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

var client *http.Client = &http.Client{}

//Gets all mail in a given inbox or outbox
func getAllBox(w http.ResponseWriter, r *http.Request) {
	boxType := mux.Vars(r)["box"]
	url := fmt.Sprintf("http://%s%s:8000/mail", boxType, serverIndex)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	var mails []Mail
	err = json.NewDecoder(res.Body).Decode(&mails)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(mails)
}

//Get all mail for a particular user for a given inbox or outbox
func getUserBox(w http.ResponseWriter, r *http.Request) {
	boxType := mux.Vars(r)["box"]
	params := mux.Vars(r)
	url := fmt.Sprintf("http://%s%s:8000/mail/%s", boxType, serverIndex, params["user"])

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	var mails []Mail
	err = json.NewDecoder(res.Body).Decode(&mails)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(mails)
}

//Deletes an email by ID in a given inbox or outbox
func delete(w http.ResponseWriter, r *http.Request) {
	boxType := mux.Vars(r)["box"]
	params := mux.Vars(r)
	url := fmt.Sprintf("http://%s%s:8000/mail/%s", boxType, serverIndex, params["email"])

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

}

//Receives an email from the MTA which has been receieved from another mail server
func receive(w http.ResponseWriter, r *http.Request) {
	var mail Mail
	url := fmt.Sprintf("http://inbox%s:8000/mail", serverIndex)
	_ = json.NewDecoder(r.Body).Decode(&mail)

	mail.ID = strconv.Itoa(rand.Intn(1000000))

	mailJSON, err := json.Marshal(mail)
	if err != nil {
		log.Fatalln(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(mailJSON))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()
}

//Sends an email to another client
func send(w http.ResponseWriter, r *http.Request) {
	var mail Mail
	url := fmt.Sprintf("http://outbox%s:8000/mail", serverIndex)
	_ = json.NewDecoder(r.Body).Decode(&mail)

	if strings.Contains(mail.From, "@") {
		if domain := strings.Split(mail.From, "@"); domain[1] != domainName {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ClientError{http.StatusBadRequest,
				"Domain name must be the same as the server's domain name",
				fmt.Sprintf("Change domain of %s to %s", mail.From, domainName)})
			return
		}
	}
	mail.ID = strconv.Itoa(rand.Intn(1000000))
	mailJSON, err := json.Marshal(mail)
	if err != nil {
		log.Fatalln(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(mailJSON))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()
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

	r.HandleFunc("/box/{box}", getAllBox).Methods("GET")
	r.HandleFunc("/box/{box}/{user}", getUserBox).Methods("GET")
	r.HandleFunc("/box/{box}/{email}", delete).Methods("DELETE")

	r.HandleFunc("/send", send).Methods("POST")
	r.HandleFunc("/receive", receive).Methods("POST")

	fmt.Printf("Server %s: MSA initialized!\n", serverIndex)
	log.Fatal(http.ListenAndServe(":8000", r))
}

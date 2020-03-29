package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

//Record - Pairs domains with an email server
type Record struct {
	Domain string `json:"domain"`
	Server string `json:"server"`
	Port   int    `json:"port"`
}

//Retrieves the environment variables to distinguish multiple email servers
var serverIndex string = os.Getenv("SERVER_INDEX")
var domainName string = os.Getenv("DOMAIN_NAME")

var client *http.Client = &http.Client{}

//Receives an email from another mail server
func receive(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("http://msa%s:8000/receive", serverIndex)
	var mail Mail
	_ = json.NewDecoder(r.Body).Decode(&mail)
	//could perform validation here

	mailJSON, err := json.Marshal(mail)
	if err != nil {
		log.Fatalln(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(mailJSON))
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	json.NewEncoder(w).Encode(mail)

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
	url := fmt.Sprintf("http://msa%s:8000/box/outbox", serverIndex)

	//Performs a function concurrently every 15 seconds to move mail from the outbox through to the correct destination
	go func() {
		for range time.Tick(time.Second * 15) {
			var mails []Mail
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Fatalln(err)
			}
			res, err := client.Do(req)
			if err != nil {
				log.Fatalln(err)
			}
			defer res.Body.Close()

			var submails []Mail
			err = json.NewDecoder(res.Body).Decode(&submails)
			if err != nil {
				log.Fatalln(err)
			}
			mails = append(mails, submails...)

			for _, mail := range mails {
				req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", url, mail.ID), nil)
				if err != nil {
					log.Fatalln(err)
				}
				res, err := client.Do(req)
				if err != nil {
					log.Fatalln(err)
				}
				defer res.Body.Close()

				req, err = http.NewRequest("GET", fmt.Sprintf("http://bluebook:8000/records/%s", mail.To), nil)
				if err != nil {
					log.Fatalln(err)
				}
				res, err = client.Do(req)
				if err != nil {
					log.Fatalln(err)
				}
				defer res.Body.Close()

				var record Record
				err = json.NewDecoder(res.Body).Decode(&record)
				if err != nil {
					log.Fatalln(err)
					//record not found, handle
				}

				mailJSON, err := json.Marshal(mail)
				if err != nil {
					log.Fatalln(err)
				}

				req, err = http.NewRequest(
					"POST",
					fmt.Sprintf("%s:%d/receive", record.Server, record.Port),
					bytes.NewBuffer(mailJSON))

				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					log.Fatalln(err)
				}

				res, err = client.Do(req)
				if err != nil {
					log.Fatalln(err)
				}
				defer res.Body.Close()
			}

		}
	}()

	r.Use(contentTypeMiddleware)
	r.HandleFunc("/receive", receive).Methods("POST")

	fmt.Printf("Server %s: MTA initialized!\n", serverIndex)
	log.Fatal(http.ListenAndServe(":8000", r))
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

//Record - Pairs domains with an email server
type Record struct {
	Domain string `json:"domain"`
	Server string `json:"server"`
	Port   int    `json:"port"`
}

//Lists all records
var records []Record

//Find a server address for a given domain name
func find(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	//email checks
	emailDomain := strings.Split(params["email"], "@")[1]
	for _, record := range records {
		if emailDomain == record.Domain {

			json.NewEncoder(w).Encode(record)
			return
		}
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
	r.HandleFunc("/records/{email}", find).Methods("GET")

	records = append(records, Record{"here.com", "http://mta0", 8000})
	records = append(records, Record{"there.com", "http://mta1", 8000})

	fmt.Println("Bluebook server initialized!")
	log.Fatal(http.ListenAndServe(":8000", r))
}

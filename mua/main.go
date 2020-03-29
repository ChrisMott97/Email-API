package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/manifoldco/promptui"
)

//Mail - Stores all the attributes for a given email
//Includes an ID primary key
type Mail struct {
	ID   string `json:"id"`
	To   string `json:"to"`
	From string `json:"from"`
	Body string `json:"body"`
}

var msa string
var domain string
var user string
var client *http.Client = &http.Client{}

func outbox() {
	var mails []Mail
	var stringMails []string
	req, err := http.NewRequest("GET", fmt.Sprintf("%sbox/outbox/%s", msa, user), nil)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&mails)
	if err != nil {
		log.Fatalln(err)
	}

	for _, mail := range mails {
		stringMails = append(stringMails, fmt.Sprintf("%s : %s", mail.To, mail.Body))
	}
	stringMails = append(stringMails, "Back")

	prompt := promptui.Select{
		Label: "Select an email to delete",
		Items: stringMails,
	}

	index, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch result {
	case "Back":
		mainMenu()
		return
	default:
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%sbox/outbox/%s", msa, mails[index].ID), nil)
		if err != nil {
			log.Fatalln(err)
		}
		res, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		defer res.Body.Close()
		fmt.Println("Deleted")
		outbox()
		return
	}
}

func inbox() {
	var mails []Mail
	var stringMails []string
	req, err := http.NewRequest("GET", fmt.Sprintf("%sbox/inbox/%s", msa, user), nil)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&mails)
	if err != nil {
		log.Fatalln(err)
	}

	for _, mail := range mails {
		stringMails = append(stringMails, fmt.Sprintf("%s : %s", mail.From, mail.Body))
	}
	stringMails = append(stringMails, "Back")

	prompt := promptui.Select{
		Label: "Select an email to delete",
		Items: stringMails,
	}

	index, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch result {
	case "Back":
		mainMenu()
	default:
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%sbox/inbox/%s", msa, mails[index].ID), nil)
		if err != nil {
			log.Fatalln(err)
		}
		res, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		defer res.Body.Close()
		fmt.Println("Deleted")
		inbox()
		return
	}
}

func send() {
	prompt := promptui.Prompt{
		Label:   "Recipient Email",
		Default: "david@there.com",
	}
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	mail := Mail{From: fmt.Sprintf("%s@%s", user, domain), To: result}

	prompt = promptui.Prompt{
		Label:   "Body",
		Default: "Hello, World!",
	}
	result, err = prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	mail.Body = result
	mailJSON, err := json.Marshal(mail)

	req, err := http.NewRequest("POST", fmt.Sprintf("%ssend", msa), bytes.NewBuffer(mailJSON))
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()
	fmt.Println("Sent!")
	mainMenu()
}

func mainMenu() {
	prompt := promptui.Select{
		Label: "Main Menu",
		Items: []string{"Outbox", "Inbox", "Send", "Logout"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch result {
	case "Outbox":
		outbox()
	case "Inbox":
		inbox()
	case "Send":
		send()
	case "Logout":
		main()
	}

}

func main() {
	var users []string
	fmt.Println("Please ensure the servers have been started with 'docker-compose up' !")
	prompt := promptui.Select{
		Label: "Select Email Server",
		Items: []string{"here.com", "there.com"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch result {
	case "here.com":
		msa = "http://localhost:8000/"
		domain = "here.com"
	case "there.com":
		msa = "http://localhost:8001/"
		domain = "there.com"
	}

	if domain == "here.com" {
		users = []string{"chris", "dan"}
	} else if domain == "there.com" {
		users = []string{"david"}
	}
	prompt = promptui.Select{
		Label: "Select a user",
		Items: users,
	}

	_, result, err = prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	user = result

	fmt.Printf("Welcome! Your email is %s@%s\n", result, domain)
	mainMenu()

}

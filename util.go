package main

import (
	"fmt"
	"log"

	"strings"
)

var defaultPrompt = "Please type (y/Y)es or (n/N)o and then press enter:"

func askForConfirmation(question string) bool {
	fmt.Println(question)
	fmt.Println(defaultPrompt)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	if len(response) == 0 {
		fmt.Println(defaultPrompt)
		return askForConfirmation(question)
	}
	if strings.ToLower(string(response[0])) == "y" {
		return true
	} else if strings.ToLower(string(response[0])) == "n" {
		return false
	} else {
		fmt.Println(defaultPrompt)
		return askForConfirmation(question)
	}
}

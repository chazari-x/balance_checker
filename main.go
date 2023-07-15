package main

import (
	"log"

	"balance_checker/cmd"
)

func main() {
	err := cmd.Start()
	if err != nil {
		log.Print(err)
	}
}

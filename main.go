package main

import (
	"log"
	"time"

	"balance_checker/cmd"
)

func main() {
	start := time.Now()
	cmd.Execute()
	log.Printf("working time: %f sec", time.Since(start).Seconds())
}

package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("Got it!")
		time.Sleep(10 * time.Second)
	}
}

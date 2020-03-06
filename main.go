package main

import (
	"fmt"
	"log"
	"os"

	"bitbucket.org/SeheonKim/albatros4/cmd"
)

func main() {
	cmd.Execute()
	//sched.Test()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
	fmt.Println("Current Working Directory: ", dir)

}

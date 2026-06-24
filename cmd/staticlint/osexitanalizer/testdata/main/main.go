package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("main called")
	os.Exit(123) // want "direct call to os.Exit in main function is forbidden"
}

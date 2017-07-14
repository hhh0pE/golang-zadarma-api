package main

import (
	"fmt"
)

func main() {
	api := APIClient{Key: "Key", Secret: "Secret"}
	fmt.Println(api.SIMs())
	fmt.Println(api.DirectNumbers())
}

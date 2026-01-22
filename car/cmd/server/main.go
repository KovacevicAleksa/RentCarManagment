package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Car servis radi!")
	http.ListenAndServe(":8090", nil)
}

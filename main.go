package main

import (
	"net/http"
	"fmt"
)



func main() {
	server()
}

func server(){
	fileServer := http.FileServer(http.Dir("./html"))
	http.Handle("/", fileServer)

	fmt.Println("clique sur le lien http://localhost:7000/")
	if err := http.ListenAndServe(":7000", nil); err != nil {
		panic(err)
	}

}
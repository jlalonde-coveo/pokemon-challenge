package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/FillIndex", PushAPI)       // this endpoint requires a get param access_token
	http.HandleFunc("/FillIndexS3", PushAPIS3)   // this endpoint requires a get param access_token
	http.HandleFunc("/pokedex", PokedexEndpoint) // development endpoint to see what a pokemon data structure looks like
	log.Fatal(http.ListenAndServe(":8888", nil))
}

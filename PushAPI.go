package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	PushSingleDataUrl = "https://apiqa.cloud.coveo.com/push/v1/organizations/pokemonchallengejasmyn04mxsqe5/sources/pokemonchallengejasmyn04mxsqe5-xi2udyjvwrglcff6ce265f2aam/documents?documentId="
)

type SinglePokemonData struct {
	Data string `json:"data"`
}

func PushAPI(writer http.ResponseWriter, request *http.Request) {
	// retrieve the pokedex
	var pokedex Pokedex
	worked := pokedex.RetrievePokedex()
	if !worked {
		WriteError(writer, ErrorPokedexRetrievalFailed)
		return
	}

	// retrieve the accessToken
	accessToken := request.URL.Query().Get("access_token")
	if accessToken == "" {
		WriteError(writer, ErrorNoAccessToken)
		return
	}

	// initialize http client
	client := &http.Client{}

	//send request for all pokemon
	var missingPokemon []Pokemon
	for _, pokemon := range pokedex.Data {

		//generate uri
		uri, err := url.Parse(fmt.Sprintf("%s%s", PokemonDbUrl, pokemon.Endpoint))
		if err != nil {
			log.Println(err)
			missingPokemon = append(missingPokemon, pokemon)
			continue
		}

		//generate the json
		data := SinglePokemonData{
			Data: pokemon.HtmlInfos,
		}
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			missingPokemon = append(missingPokemon, pokemon)
			continue
		}

		// generate put request
		pushRequest, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s%s", PushSingleDataUrl, uri), bytes.NewReader(jsonBytes))
		if err != nil {
			log.Println(err)
			missingPokemon = append(missingPokemon, pokemon)
			continue
		}
		SetHeaders(pushRequest, accessToken)

		// send put request
		pushResponse, err := client.Do(pushRequest)
		if err != nil {
			log.Println(err)
			missingPokemon = append(missingPokemon, pokemon)
			continue
		}

		//analyze response
		if pushResponse.StatusCode != 202 {
			pushResponseBody, err := ioutil.ReadAll(pushResponse.Body)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("error : %d with body %s", pushResponse.StatusCode, string(pushResponseBody))
			}
			missingPokemon = append(missingPokemon, pokemon)
			continue
		}

		pushResponse.Body.Close()
		time.Sleep(5 * time.Millisecond)
	}

	if len(missingPokemon) != 0 {
		_, err := writer.Write([]byte("The following pokemon documents couldn't be pushed :\n"))
		if err != nil {
			log.Println(err)
		}
		for _, pokemon := range missingPokemon {
			_, err := writer.Write(pokemon.ToJSON())
			if err != nil {
				log.Println(err)
			}
			_, err = writer.Write([]byte("\n"))
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		_, err := writer.Write([]byte("{}"))
		if err != nil {
			log.Println(err)
		}
	}
}

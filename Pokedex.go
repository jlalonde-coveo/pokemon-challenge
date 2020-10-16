package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
)

const (
	PokemonDbUrl    = "https://pokemondb.net"
	GeneralEndpoint = "/pokedex/national"
)

type Pokemon struct {
	Name      string
	Endpoint  string
	HtmlInfos string
}

type Pokedex struct {
	Data []Pokemon `json:"data"`
}

func (pokedex *Pokedex) RetrievePokedex() (worked bool) {
	// Request the HTML page
	response, err := http.Get(PokemonDbUrl + GeneralEndpoint)
	if err != nil {
		log.Print(err)
		return false
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		log.Printf("status code error: %d %s", response.StatusCode, response.Status)
		return false
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Print(err)
		return false
	}

	// get the infos
	doc.Find(".infocard").Find(".infocard-lg-data.text-muted").Each(func(_ int, selection *goquery.Selection) {
		//store in pokedex
		node := selection.Find("a").Nodes[0]
		pokedex.Add(Pokemon{
			Name:     node.FirstChild.Data,
			Endpoint: node.Attr[1].Val,
		})
	})

	// access specific pokemon pages
	for i, pokemon := range pokedex.Data {
		//request
		response, err = http.Get(PokemonDbUrl + pokemon.Endpoint)
		if err != nil {
			log.Print(err)
			return false
		}
		if response.StatusCode != 200 {
			log.Printf("status code error: %d %s", response.StatusCode, response.Status)
			return false
		}

		// Load the HTML document
		doc, err = goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Print(err)
			return false
		}

		// get the infos
		html, err := doc.Find("#main").Html()
		if err != nil {
			log.Print(err)
			return false
		}
		pokedex.Data[i].HtmlInfos = html

		response.Body.Close()
	}
	return true
}

func (pokedex *Pokedex) Add(pokemon Pokemon) {
	pokedex.Data = append(pokedex.Data, pokemon)
}

func (pokemon *Pokemon) ToJSON() (jsonBytes []byte) {
	type Data struct {
		Data string `json:"data"`
	}
	data := Data{
		Data: pokemon.Name,
	}
	jsonBytes, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		log.Print(err)
	}
	return jsonBytes
}

func PokedexEndpoint(writer http.ResponseWriter, _ *http.Request) {
	var pokedex Pokedex
	worked := pokedex.RetrievePokedex()
	if !worked {
		log.Println("Couldn't retrieve pokedex")
	}
	jsonBytes, err := json.MarshalIndent(pokedex, "", "	")
	if err != nil {
		log.Println(err)
	}
	_, err = writer.Write(jsonBytes)
	if err != nil {
		log.Println(err)
	}
}

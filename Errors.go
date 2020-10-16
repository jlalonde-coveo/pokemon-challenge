package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	ErrorPokedexRetrievalFailed = "Failed to retrieve pokedex"
	ErrorNoAccessToken          = "No access_token parameter were given"
	ErrorCreatingS3             = "Couldn't create the s3 container"
	ErrorCreatingS3Payload      = "Couldn't create the s3 payload"
	ErrorPushingToS3            = "Couldn't push to s3 bucket"
	ErrorPushingToIndex         = "Couldn't push to index"
)

func WriteError(writer http.ResponseWriter, errorMessage string) {
	type ErrorMessage struct {
		Message string
	}
	message := ErrorMessage{
		Message: errorMessage,
	}
	jsonBytes, err := json.MarshalIndent(message, "", "	")
	if err != nil {
		log.Println(err)
	}
	_, err = writer.Write(jsonBytes)
	if err != nil {
		log.Println(err)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	S3Url         = "https://apiqa.cloud.coveo.com/push/v1/organizations/pokemonchallengejasmyn04mxsqe5/files"
	PushFromS3Url = "https://apiqa.cloud.coveo.com/push/v1/organizations/pokemonchallengejasmyn04mxsqe5/sources/pokemonchallengejasmyn04mxsqe5-xi2udyjvwrglcff6ce265f2aam/documents/batch?fileId="
)

type S3Container struct {
	UploadUri       string            `json:"uploadUri"`
	FileId          string            `json:"fileId"`
	RequiredHeaders map[string]string `json:"requiredHeaders"`
}

type S3Payload struct {
	AddOrUpdate []AddOrUpdateItem `json:"addOrUpdate"`
	Delete      []DeleteItem      `json:"delete"`
}

type AddOrUpdateItem struct {
	DocumentId    string
	Data          string
	FileExtension string
}

type DeleteItem struct {
	DocumentId     string `json:"documentId"`
	DeleteChildren bool   `json:"deleteChildren"`
}

func PushAPIS3(writer http.ResponseWriter, request *http.Request) {
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

	// create the s3 bucket
	s3Container, err := CreateS3Container(client, accessToken)
	if err != nil || s3Container == nil {
		WriteError(writer, ErrorCreatingS3)
		return
	}

	// create the s3 payload
	s3Payload, err := CreateS3Payload(pokedex)
	if err != nil {
		WriteError(writer, ErrorCreatingS3Payload)
		return
	}

	// push s3 payload to s3
	worked, err = PushToS3(client, s3Container, s3Payload)
	if !worked {
		if err != nil {
			log.Println(err)
		}
		WriteError(writer, ErrorPushingToS3)
		return
	}

	// push the s3 container to the push source
	worked, err = PushToIndex(client, accessToken, s3Container.FileId)
	if !worked {
		if err != nil {
			log.Println(err)
		}
		WriteError(writer, ErrorPushingToIndex)
		return
	}

	_, err = writer.Write([]byte("{}"))
	if err != nil {
		log.Println(err)
	}
}

func CreateS3Container(client *http.Client, accessToken string) (s3Container *S3Container, err error) {
	// set the HTTP method, url, and request body
	s3Request, err := http.NewRequest(http.MethodPost, S3Url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, err
	}
	SetHeaders(s3Request, accessToken)

	// execute the request
	s3Response, err := client.Do(s3Request)
	if err != nil {
		return nil, err
	}

	// retrieve the body
	s3ResponseBody, err := ioutil.ReadAll(s3Response.Body)
	if err != nil {
		return nil, err
	}
	s3Response.Body.Close()

	// validate return code
	if s3Response.StatusCode != 201 {
		log.Printf("error : %d with body %s", s3Response.StatusCode, string(s3ResponseBody))
		return nil, nil
	}

	//create the json
	err = json.Unmarshal(s3ResponseBody, &s3Container)
	if err != nil {
		return nil, err
	}
	return s3Container, nil
}

func CreateS3Payload(pokedex Pokedex) (s3Payload *S3Payload, err error) {
	//create the json
	var addOrUpdate []AddOrUpdateItem
	for _, pokemon := range pokedex.Data {
		//generate uri
		uri, err := url.Parse(fmt.Sprintf("%s%s", PokemonDbUrl, pokemon.Endpoint))
		if err != nil {
			return nil, err
		}

		//create the add item
		addOrUpdateItem := AddOrUpdateItem{
			DocumentId:    uri.RequestURI(),
			Data:          pokemon.HtmlInfos,
			FileExtension: ".html",
		}
		addOrUpdate = append(addOrUpdate, addOrUpdateItem)
	}
	s3Payload = &S3Payload{
		AddOrUpdate: addOrUpdate,
	}
	return s3Payload, err
}

func PushToS3(client *http.Client, s3Container *S3Container, s3Payload *S3Payload) (worked bool, err error) {

	// convert payload into json
	jsonBytes, err := json.Marshal(s3Payload)
	if err != nil {
		return false, err
	}

	// generate put request
	s3Request, err := http.NewRequest(http.MethodPut, s3Container.UploadUri, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return false, err
	}
	for k, v := range s3Container.RequiredHeaders {
		s3Request.Header.Add(k, v)
	}

	log.Println(bytes.NewBuffer(jsonBytes))

	// send put request
	s3Response, err := client.Do(s3Request)
	if err != nil {
		return false, err
	}

	// analyze response
	if s3Response.StatusCode != 200 {
		s3ResponseBody, err := ioutil.ReadAll(s3Response.Body)
		if err != nil {
			return false, err
		}
		log.Printf("error : %d with body %s", s3Response.StatusCode, string(s3ResponseBody))
		return false, nil
	}

	s3Response.Body.Close()
	return true, nil

}

func PushToIndex(client *http.Client, accessToken string, fileId string) (worked bool, err error) {
	// generate put request
	pushRequest, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s%s", PushFromS3Url, fileId), bytes.NewReader([]byte("{}")))
	if err != nil {
		return false, err
	}
	SetHeaders(pushRequest, accessToken)

	// send put request
	pushResponse, err := client.Do(pushRequest)
	if err != nil {
		return false, err
	}

	// analyze response
	if pushResponse.StatusCode != 202 {
		pushResponseBody, err := ioutil.ReadAll(pushResponse.Body)
		if err != nil {
			return false, err
		}
		log.Printf("error : %d with body %s", pushResponse.StatusCode, string(pushResponseBody))
		return false, nil
	}

	pushResponse.Body.Close()
	return true, nil
}

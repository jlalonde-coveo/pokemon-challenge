package main

import (
	"fmt"
	"net/http"
)

func SetHeaders(request *http.Request, accessToken string) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
}

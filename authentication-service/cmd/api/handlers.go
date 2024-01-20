package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *Config) Authenticate(rw http.ResponseWriter, r *http.Request) {

	var requestPayload struct {
		Email    string `json:"email"`
		Password string `jsom:"password"`
	}

	err := app.readJSON(rw, r, &requestPayload)

	if err != nil {
		app.errorJSON(rw, err, http.StatusBadRequest)
		return
	}

	//validate the user against db
	user, err := app.Models.User.GetByEmail(requestPayload.Email)

	if err != nil {
		app.errorJSON(rw, errors.New("invalid Credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)

	if err != nil || !valid {
		app.errorJSON(rw, errors.New("Password is not valid"), http.StatusBadRequest)
		return
	}

	err = app.logRequest(rw, "authentication", fmt.Sprintf("Authentication is established with user %s", requestPayload.Email))
	if err != nil {
		app.errorJSON(rw, errors.New("Error While Logging Authentication"), http.StatusInternalServerError)
		log.Panic(err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in User %s", requestPayload.Email),
		Data:    user,
	}

	app.writeJSON(rw, http.StatusAccepted, payload)
}

func (app *Config) logRequest(rw http.ResponseWriter, name string, data string) error {

	var logPayload struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	logPayload.Data = data
	logPayload.Name = name

	jsonPayload, _ := json.MarshalIndent(logPayload, "", "\t")

	loggerServiceURL := "http://logger-service/log"
	loggerRequest, err := http.NewRequest("POST", loggerServiceURL, bytes.NewBuffer(jsonPayload))

	if err != nil {
		app.errorJSON(rw, err)
		return err
	}

	client := &http.Client{}

	_, err = client.Do(loggerRequest)

	if err != nil {
		app.errorJSON(rw, err)
		return err
	}

	return nil

}

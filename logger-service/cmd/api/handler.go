package main

import (
	"log"
	"log-service/data"
	"net/http"
)

type JSonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(rw http.ResponseWriter, r *http.Request) {

	var requestPayload JSonPayload
	_ = app.readJSON(rw, r, &requestPayload)

	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	log.Println("In the Log Service Handler")
	err := app.Models.LogEntry.Insert(event)
	log.Println("After Insert")
	if err != nil {
		log.Println("Error Occured", err.Error())
		app.errorJSON(rw, err)
		return
	}
	resp := JsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJSON(rw, http.StatusAccepted, resp)
}

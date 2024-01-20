package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string        `json:"action"`
	Auth   AuthPayload   `json:"auth,omitempty"`
	Log    LoggerPayload `json:"log,omitempty"`
	Mail   MailPayload   `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoggerPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(wr http.ResponseWriter, r *http.Request) {

	responsePayload := JsonResponse{
		Error:   false,
		Message: "Broker Service Called",
	}

	_ = app.writeJSON(wr, http.StatusOK, responsePayload)
}

func (app *Config) HandleSubmission(wr http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJSON(wr, r, &requestPayload)
	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(wr, requestPayload.Auth)
	case "log":
		app.logItemViaRPC(wr, requestPayload.Log)
		//app.LogItem(wr, requestPayload.Log)
		//app.logEventviaRabbit(wr, requestPayload.Log)
	case "mail":
		app.SendMail(wr, requestPayload.Mail)
	default:
		app.errorJSON(wr, errors.New("unknown action "))

	}

}

func (app *Config) LogItem(wr http.ResponseWriter, entry LoggerPayload) {
	//only in development, production : use Indent
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(wr, err)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Logged"

	app.writeJSON(wr, http.StatusAccepted, payload)

}

func (app *Config) authenticate(wr http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(wr, errors.New("Invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(wr, errors.New("Unexpected Error thrown in auth service"))
		return
	}

	var jsonfromService JsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonfromService)
	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	if jsonfromService.Error {
		app.errorJSON(wr, err, http.StatusUnauthorized)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonfromService.Data

	app.writeJSON(wr, http.StatusAccepted, payload)
}

func (app *Config) SendMail(wr http.ResponseWriter, m MailPayload) {

	jsonData, _ := json.MarshalIndent(m, "", "\t")
	request, err := http.NewRequest("POST", "http://mail-service/send", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(wr, errors.New("Unexpected Error thrown in mail service"))
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Message Sent to" + m.To

	app.writeJSON(wr, http.StatusAccepted, payload)
}

func (app *Config) logEventviaRabbit(wr http.ResponseWriter, l LoggerPayload) {

	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(wr, err)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Logged via RabbitMQ"

	app.writeJSON(wr, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error {

	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LoggerPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	emitter.Push(string(j), "log.INFO")

	if err != nil {
		return err
	}
	return nil
}

// Should be in exact name with the RPC server
type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(wr http.ResponseWriter, lp LoggerPayload) {

	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		log.Println("Error connecting to RPC server")
		app.errorJSON(wr, err)
		log.Println(err)
		return
	}

	rpcPayload := RPCPayload{
		Name: lp.Name,
		Data: lp.Data,
	}
	var result string

	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		log.Println("Error calling RPC server")
		app.errorJSON(wr, err)
		log.Println(err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(wr, http.StatusAccepted, payload)
}

func (app *Config) logItemViaGRPC(wr http.ResponseWriter, r *http.Request) {

	var requestPayload RequestPayload

	err := app.readJSON(wr, r, &requestPayload)

	if err != nil {
		log.Println("Error while reading Request json")
		app.errorJSON(wr, err)
		return
	}

	//while connecting to gRPC server, you need to provide the credentials
	//to connect to gRPC server insecurely (localhost) with no security credentials
	//if connection established successfully, it returns a new grpc.ClientConnection
	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

	if err != nil {
		log.Println("Error while connecting to gRPC server")
		app.errorJSON(wr, err)
		return
	}
	//close connection when done
	defer conn.Close()

	//create new gRPC client with connection to gRPC server
	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Logs{Name: requestPayload.Log.Name, Data: requestPayload.Log.Data},
	}, grpc.WaitForReady(true))

	if err != nil {
		log.Println("Error while calling gRPC server")
		app.errorJSON(wr, err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: "logged via gRPC",
	}

	app.writeJSON(wr, http.StatusAccepted, payload)
}

package main

import (
	"authentication/events"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type LogPayload struct{
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Auth(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJson(w,r,&requestPayload)
	if err != nil {
		app.errorJson(w, err, http.StatusBadRequest)
		return
	}

	//validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.errorJson(w, errors.New("invalid email"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	//log auth
	// err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	// if err != nil {
	// 	app.errorJson(w, err)
	// 	return
	// }

	//log auth using rabbitMQ
	err = app.logRequestViaRabbitMQ(w,"authentication", fmt.Sprintf("%s logged in via RabbitMQ", user.Email))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := JsonResponse{
		Error: false,
		Message: fmt.Sprintf("logged in user %s", user.Email),
		Data: user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequestViaRabbitMQ(w http.ResponseWriter, name string, message string) error {
	err := app.pushToQueue(name,message)
	
	if err != nil {
		// app.errorJson(w, err)
		return err
	}

	return nil

	// var payload JsonResponse
	// payload.Error = false
	// payload.Message = "Auth via RabbitMQ"

	// app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name string, message string) error {
	emitter, err := events.NewEvenEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return err
}

func (app *Config) logRequest(name , data string) error {
	var entry struct{
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "","\t")
	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := http.Client{}

	_,err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
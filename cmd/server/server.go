package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ReqObj struct {
	Handle    string
	Passcode  string
	Timestamp int
	TicketId  int
	TFrom     string
	TTo       string
}

type EventObj struct {
	event string
	id    int
	time  int
}
type SyncObj struct {
	client_timestamp int
	events           []EventObj
}

const chkUrl = "https://checkin.timepad.ru"

func main() {

	http.Handle("/", http.FileServer(http.Dir("../../assets")))

	http.HandleFunc("/req", func(w http.ResponseWriter, r *http.Request) {

		// Allow requests from different origins
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// This API responds only with JSON, unlike Timepad
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost {

			// Parse input
			var obj ReqObj
			err := json.NewDecoder(r.Body).Decode(&obj)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Do a request to Timepad API
			resp, err := apiReq(obj)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			fmt.Fprint(w, resp)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":3250", nil))

}

// apiReq does the actual request to Timepad API and gets the response
func apiReq(request ReqObj) (string, error) {
	if request.Handle == "" {
		return "", errors.New("no Handle parameter provided")
	}

	switch request.Handle {
	case "/login":
		// Prepare proxy vars
		data := url.Values{
			"code":             {request.Passcode},
			"client_timestamp": {fmt.Sprint(request.Timestamp)},
		}

		req, err := http.NewRequest("POST", chkUrl+request.Handle, strings.NewReader(data.Encode()))

		if err != nil {
			log.Fatal(err)
			return "", errors.New("failed to construct request for timepad")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatal(err)
			return "", errors.New("failed to post to timepad")
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return buf.String(), nil

	case "/sync":
		// It is unclear what is the difference between time and client_timestamp,
		// so they are both populated with the client timestamp
		event := EventObj{"checked_in", request.TicketId, request.Timestamp}
		data := SyncObj{request.Timestamp, []EventObj{event}}
		b, err := json.Marshal(data)

		if err != nil {
			return "", errors.New("failed to construct JSON from input")
		}

		req, err := http.NewRequest("POST", chkUrl+request.Handle, bytes.NewBuffer(b))
		if err != nil {
			log.Fatal(err)
			return "", errors.New("failed to construct request for timepad")
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
			return "", errors.New("failed to post to timepad")
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return buf.String(), nil

	default:
		return "", errors.New("handle is not allowed")
	}

}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ReqObj struct {
	Handle    string
	Passcode  string
	Timestamp string
	TicketId  string
	TFrom     string
	TTo       string
}

const chkUrl = "https://checkin.timepad.ru"

func main() {

	http.Handle("/", http.FileServer(http.Dir("../../assets")))

	http.HandleFunc("/req", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprint(w, "Not yet implemented")
		case http.MethodPost:

			// Parse input
			var obj ReqObj
			err := json.NewDecoder(r.Body).Decode(&obj)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Prepare proxy vars
			data := url.Values{
				"code":             {obj.Passcode},
				"client_timestamp": {obj.Timestamp},
			}

			req, err := http.NewRequest("POST", chkUrl+obj.Handle, strings.NewReader(data.Encode()))

			if err != nil {
				log.Fatal(err)
				fmt.Fprint(w, "{\"error\":\"Failed to POST to timepad\"}")
				return
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := http.DefaultClient.Do(req)

			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			fmt.Fprint(w, buf.String())

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":3250", nil))

}

func apiReq() {

}

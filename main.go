package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var envVarNames = []string{"TWILIO_NUMBER", "TARGET_NUMBER", "VOICEMAIL_EMAIL"}

var envVars map[string]string

func main() {
	addr := os.Getenv("BIND_ADDRESS")
	if addr == "" {
		addr = "localhost:8080"
	}
	envVars = make(map[string]string, len(envVarNames))
	for _, name := range envVarNames {
		val := os.Getenv(name)
		if val != "" {
			envVars[name] = val
		}
	}
	log.Println(fmt.Sprintf("About to start listen on %s", addr))
	log.Fatal(http.ListenAndServe(addr, handler()))
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/incoming-call", StraightToVoiceMail)
	mux.HandleFunc("/process-recording", HandleRecordCallback)
	return mux
}

func sendTwiML(w http.ResponseWriter, twiML string) {
	tmpl, err := template.New("twiMLResponse").Parse(twiML)
	if err != nil {
		log.Printf("error creating TwiML template: %v\n", err)
	}

	w.Header().Add("Content-Type", "text/xml; charset=utf-8")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))

	err = tmpl.Execute(w, envVars)
	if err != nil {
		log.Printf("error while executing TwiML template: %v\n", err)
	}
}

// StraightToVoiceMail uses twimlet to record a message and send it over by email
func StraightToVoiceMail(w http.ResponseWriter, _ *http.Request) {
	xml := `
<Response>
    <Redirect method="GET">http://twimlets.com/voicemail?Email={{ .VOICEMAIL_EMAIL }}&amp;Message=Please+Leave+A+Message</Redirect>
</Response>
`
	sendTwiML(w, xml)
}

// HandleIncomingCall serves XML to kick off incoming call
func HandleIncomingCall(w http.ResponseWriter, req *http.Request) {
	responseFormat := `
<Response>
    <Say voice="alice" language="en-US">Please leave a message</Say>
    <Sms from="{{ .TWILIO_NUMBER }}" to="{{ .TARGET_NUMBER }}">Call from %s, %s, %s.</Sms>
    <Record maxLength="300" action="/process-recording"></Record>
    <Say voice="alice" language="en-US">I may not give you a call back. Please leave a message.</Say>
    <Record maxLength="300" action="/process-recording"></Record>
    <Say voice="alice" language="en-US">Ok, as you wish, good bye</Say>
</Response>
`
	city := getRequestParam(req, "FromCity", "Unknown city")
	fromNumber := getRequestParam(req, "From", "Unknown number")
	country := getRequestParam(req, "FromCountry", "Unknown country")

	xml := fmt.Sprintf(responseFormat, city, fromNumber, country)
	sendTwiML(w, xml)
}

// HandleRecordCallback handles Twilio callback with a recorded message
func HandleRecordCallback(w http.ResponseWriter, req *http.Request) {
	responseFormat := `
<Response>
	<Say voice="alice" language="en-US">Thank you, you should hear back soon</Say>
	<Sms from="{{ .TWILIO_NUMBER }}" to="{{ .TARGET_NUMBER }}">You have got a twilio call from %s, %s, %s</Sms>
</Response>
	`
	if err := req.ParseForm(); err != nil {
		log.Printf("Failed to parse form %v\n", err)
	}
	city := getRequestParam(req, "FromCity", "Unknown city")
	fromNumber := getRequestParam(req, "From", "Unknown number")
	country := getRequestParam(req, "FromCountry", "Unknown country")
	_ = getRequestParam(req, "RecordingUrl", "no message")

	xml := fmt.Sprintf(responseFormat, city, fromNumber, country)
	sendTwiML(w, xml)
}

func getRequestParam(req *http.Request, paramName, defaultValue string) string {
	value := paramGet(req, paramName)
	if value == "" && defaultValue != "" {
		return defaultValue
	}

	return value
}

func paramGet(req *http.Request, key string) string {
	if req.Method == "GET" {
		return req.URL.Query().Get(key)
	}
	return req.Form.Get(key)
}

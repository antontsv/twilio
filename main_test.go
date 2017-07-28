package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetHandlerBasicResponse(t *testing.T) {
	tt := []struct {
		endpoint   string
		statusCode int
	}{
		{endpoint: "incoming-call", statusCode: http.StatusOK},
		{endpoint: "process-recording", statusCode: http.StatusOK},
		{endpoint: "incoming-call-experimental-flow", statusCode: http.StatusOK},
	}
	srv := httptest.NewServer(handler())
	defer srv.Close()
	for _, tc := range tt {
		t.Run(tc.endpoint, func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/%s", srv.URL, tc.endpoint))
			if err != nil {
				t.Fatalf("could not send request to %q", tc.endpoint)
			}
			if status := resp.StatusCode; status != tc.statusCode {
				t.Fatalf("returned wrong status code: got %v expected %v", status, tc.statusCode)
			}
			defer resp.Body.Close()
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("could not read body of %q", tc.endpoint)
			}
			if err := checkWellFormedXML(string(bytes)); err != nil {
				t.Errorf("body needs to be valid XML: %v", err)
			}
		},
		)
	}
}

func checkWellFormedXML(s string) error {
	d := xml.NewDecoder(strings.NewReader(s))
	t, err := d.Token()
	if err != nil {
		return err
	}

	v, ok := t.(xml.ProcInst)
	if !ok || v.Target != "xml" || !strings.Contains(string(v.Inst), "version=\"1.0\"") {
		return fmt.Errorf("No XML header detected with version 1.0 at the start")
	}

	return nil
}

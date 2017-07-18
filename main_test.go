package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStraingToVoiceMailResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	StraightToVoiceMail(recorder, nil)
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("returned wrong status code: got %v expected %v", status, http.StatusOK)
	}
	if err := checkWellFormetXML(recorder.Body.String()); err != nil {
		t.Errorf("body need to be valid XML: %v", err)
	}
}

func checkWellFormetXML(s string) error {
	if !strings.HasPrefix(s, "<?xml ") {
		return fmt.Errorf("No XML header detected at the start")
	}
	d := xml.NewDecoder(strings.NewReader(s))
	_, err := d.Token()
	return err
}

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r),nil
}

func NewtestClient(fn RoundTripFunc) *http.Client{
	return &http.Client{
		Transport: fn,
	}
}

func TestAuth(t *testing.T) {
	jsonToReturn := `
{
	"error": false,
	"message": "some message"
}		
`

	client := NewtestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(bytes.NewBufferString(jsonToReturn)),
			Header: make(http.Header),
		}
	})

	testApp.Client = client

	postBody := map[string]interface{}{
		"email": "me@here.com",
		"password": "verysecret",
	}

	body,_ := json.Marshal(postBody)

	req,_ := http.NewRequest("POST", "/auth", bytes.NewReader(body))
	responseRecord := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.Auth)

	handler.ServeHTTP(responseRecord, req)

	if responseRecord.Code != http.StatusAccepted{
		t.Errorf("EXPECTED HTTP STATUS ACCEPTED BUT GOT %d", responseRecord.Code)
	}
}
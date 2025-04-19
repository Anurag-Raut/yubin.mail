package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestMail(t *testing.T) {

	clientServer := NewClientServer("127.0.0.1", "8001")
	go clientServer.Listen()

	time.Sleep(1 * time.Second)
	mailBody := map[string]any{
		"from": "test@localhost",
		"to":   []string{"receiver@localhost"},
		"body": "This is a test email",
	}
	bodyBytes, err := json.Marshal(mailBody)

	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post("http://127.0.0.1:8001/newRequest", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatal("FAILED to sened POST REQ")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 got %d", res.StatusCode)
	}
	resBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(resBodyBytes))
	t.Cleanup(func() {

		clientServer.Close()
	})

}

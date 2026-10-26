//go:build !js

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIHealthSolvePanel(t *testing.T) {
	net := LoadMLP(dataPath("model_math.json"))
	if net == nil {
		t.Fatal("modelo embebido no carga")
	}
	srv := httptest.NewServer(apiHandler(net))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("/health status=%d", res.StatusCode)
	}
	res.Body.Close()

	body := `{"a":{"op":"add","a":{"op":"var","s":"a"},"b":{"op":"const","n":0}},"b":{"op":"var","s":"a"},"cap":100}`
	res, err = http.Post(srv.URL+"/solve", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	var sol solveResponse
	json.NewDecoder(res.Body).Decode(&sol)
	res.Body.Close()
	if res.StatusCode != 200 || !sol.Ok {
		t.Fatalf("/solve status=%d ok=%v error=%q", res.StatusCode, sol.Ok, sol.Error)
	}

	res, err = http.Get(srv.URL + "/panel?cap=2")
	if err != nil {
		t.Fatal(err)
	}
	var panel panelResult
	json.NewDecoder(res.Body).Decode(&panel)
	res.Body.Close()
	if res.StatusCode != 200 || panel.Problemas != 700 || panel.Cap != 2 {
		t.Fatalf("/panel status=%d problemas=%d cap=%d", res.StatusCode, panel.Problemas, panel.Cap)
	}
}

func TestAPIMetodosYJSONInvalidos(t *testing.T) {
	srv := httptest.NewServer(apiHandler(LoadMLP(dataPath("model_math.json"))))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/solve")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /solve status=%d", res.StatusCode)
	}
	res.Body.Close()

	res, err = http.Post(srv.URL+"/solve", "application/json", strings.NewReader("{"))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /solve JSON roto status=%d", res.StatusCode)
	}
	res.Body.Close()
}

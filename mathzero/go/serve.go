//go:build !js

// API local minima (Q8): expone el motor por HTTP para poder probarlo con
// curl sin escribir un cliente. Pensada para localhost, sin autenticacion.
//
//	mathzero serve [addr]
//	GET  /health
//	GET  /panel?cap=1000
//	POST /solve  {"a":{...},"b":{...},"cap":1000}
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var apiMu sync.Mutex

type solveRequest struct {
	A   json.RawMessage `json:"a"`
	B   json.RawMessage `json:"b"`
	Cap int             `json:"cap"`
}

type solveResponse struct {
	Ok    bool   `json:"ok"`
	Nodos int    `json:"nodos"`
	Pasos int    `json:"pasos"`
	Error string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func apiHandler(net *MLP) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "usa GET"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "nivel": net.Level, "ronda": net.Round,
			"endpoints": []string{"/health", "/panel?cap=1000", "/solve"},
		})
	})
	mux.HandleFunc("/panel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "usa GET"})
			return
		}
		cap := 1000
		if raw := r.URL.Query().Get("cap"); raw != "" {
			if v, ok := parseIntArg("cap", raw, 1); ok {
				cap = v
			} else {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cap invalido"})
				return
			}
		}
		apiMu.Lock()
		defer apiMu.Unlock()
		DirFix = true
		LemmaLib = nil
		lv := map[int][]MState{}
		for k := 2; k <= 8; k++ {
			lv[k] = LoadProblems(dataPath("blind.json"), fmt.Sprint(k))
		}
		res := evalPanel(net, lv, cap)
		res.Modelo, res.Nivel, res.Ronda, res.Cap = "model_math.json", net.Level, net.Round, cap
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("/solve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "usa POST"})
			return
		}
		var req solveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalido"})
			return
		}
		if len(req.A) == 0 || len(req.B) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "faltan a o b"})
			return
		}
		cap := req.Cap
		if cap <= 0 {
			cap = 1000
		}
		var resp solveResponse
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					resp = solveResponse{Error: fmt.Sprint("expresion invalida: ", rec)}
				}
			}()
			apiMu.Lock()
			defer apiMu.Unlock()
			DirFix = true
			LemmaLib = nil
			a, b := parseExpr(req.A), parseExpr(req.B)
			ok, nodes, sm := GuidedSolve(net, a, b, cap)
			resp = solveResponse{Ok: ok, Nodos: nodes, Pasos: len(sm)}
		}()
		if resp.Error != "" {
			writeJSON(w, http.StatusBadRequest, resp)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
	return mux
}

func ServeMain(args []string) {
	rest, ok := parseDataFlags(args)
	if !ok {
		return
	}
	if len(rest) > 1 {
		fmt.Fprintln(os.Stderr, "uso: mathzero serve [addr] [-datadir DIR]")
		return
	}
	addr := "127.0.0.1:8080"
	if len(rest) == 1 {
		addr = rest[0]
	}
	net := LoadMLP(dataPath("model_math.json"))
	if net == nil {
		fmt.Fprintln(os.Stderr, "[serve] no se pudo cargar el modelo")
		return
	}
	fmt.Println("[serve] API local en http://" + addr)
	fmt.Println("[serve] GET /health | GET /panel?cap=1000 | POST /solve")
	if err := http.ListenAndServe(addr, apiHandler(net)); err != nil {
		fmt.Fprintln(os.Stderr, "[serve] error:", err)
	}
}

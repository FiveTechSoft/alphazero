//go:build !js && linux

// Tests de d1serve v3 (worker subproceso + guard.sh). Concurrencia util solo
// con -race: go test -race -run D1Serve .
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func newD1sTestServer(t *testing.T, hogArgs []string) (*httptest.Server, *d1sServer) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	sv := &d1sServer{
		model: "", capN: 50, rssMB: 96, wtimeout: 5,
		guard: d1sTestGuard(t), exe: exe,
		worker: make(chan struct{}, 1), queue: make(chan struct{}, d1sQueueSlots),
		testWorkerArgs: hogArgs,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", sv.handleModels)
	mux.HandleFunc("/v1/chat/completions", sv.handleCompletions)
	s := httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s, sv
}

// d1sTestGuard localiza guard.sh de forma reproducible: override D1S_GUARD,
// layout del repo (integrations/opencode/guard.sh desde mathzero/go) o copia
// en el directorio de trabajo (staging documentado en OPENCODE.md).
func d1sTestGuard(t *testing.T) string {
	t.Helper()
	if g := os.Getenv("D1S_GUARD"); g != "" {
		if _, err := os.Stat(g); err == nil {
			return g
		}
		t.Fatalf("D1S_GUARD no existe: %s", g)
	}
	for _, c := range []string{"../../integrations/opencode/guard.sh", "guard.sh"} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	t.Fatal("guard.sh no encontrado: define D1S_GUARD o usa el layout del repo (integrations/opencode/guard.sh)")
	return ""
}

func d1sPost(t *testing.T, url, body string) (int, string) {
	t.Helper()
	resp, err := http.Post(url+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func d1sSolveBody(problem string) string {
	return fmt.Sprintf(`{"model":"alphazero","messages":[{"role":"user","content":%q}]}`, problem)
}

func TestD1ServeConcurrentSolveRace(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	const N = 16
	var wg sync.WaitGroup
	codes := make([]int, N)
	bodies := make([]string, N)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, b := d1sPost(t, s.URL, d1sSolveBody("add(va,c0) -> va"))
			codes[i], bodies[i] = c, b
		}(i)
	}
	wg.Wait()
	ok200, rej := 0, 0
	for i := 0; i < N; i++ {
		if codes[i] == 200 && strings.Contains(bodies[i], "Resuelto") {
			ok200++
		} else if codes[i] == 503 {
			rej++
		} else {
			t.Errorf("req %d: codigo inesperado %d: %.120s", i, codes[i], bodies[i])
		}
	}
	if ok200 == 0 {
		t.Error("ninguna resolucion completada")
	}
	t.Logf("200-OK=%d 503=%d", ok200, rej)
}

func TestD1ServeMixedConcurrent(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	inputs := []string{"add(va,c0) -> va", "hola", "quien eres", "cuentame un chiste", "neg(neg(c4)) -> neg(neg(mul(neg(neg(c4)),c1)))", "hi"}
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, b := d1sPost(t, s.URL, d1sSolveBody(inputs[i%len(inputs)]))
			if c != 200 && c != 503 {
				t.Errorf("input %q: codigo %d: %.120s", inputs[i%len(inputs)], c, b)
			}
		}(i)
	}
	wg.Wait()
}

func TestD1ServeBodyTooLarge(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	big := `{"model":"alphazero","messages":[{"role":"user","content":"` + strings.Repeat("x", 300<<10) + `"}]}`
	c, _ := d1sPost(t, s.URL, big)
	if c != 413 {
		t.Errorf("esperado 413, obtenido %d", c)
	}
}

func TestD1ServeExprTooLong(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	long := strings.Repeat("neg(", 50) + "c4" + strings.Repeat(")", 50) // >200 bytes, profundidad 50
	c, b := d1sPost(t, s.URL, d1sSolveBody(long+" -> c4"))
	if c != 400 || !strings.Contains(b, "expresion") {
		t.Errorf("esperado 400 con explicacion, obtenido %d: %.160s", c, b)
	}
}

func TestD1ServeParseErrorSinNil(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	c, b := d1sPost(t, s.URL, d1sSolveBody("perro -> gato"))
	if c != 200 || !strings.Contains(b, "cannot parse") || strings.Contains(b, "<nil>") {
		t.Errorf("parseo: %d %.160s", c, b)
	}
}

func TestD1ServeRutas(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	casos := []struct{ in, want string }{
		{"hola", "Hola. No soy un modelo de lenguaje"},
		{"Hola, quien eres", "Soy AlphaZero (MathZero)"},
		{"hi", "Hello. I am not a language model"},
		{"who are you", "I am AlphaZero (MathZero)"},
		{"cuentame un chiste", "fuera de mi alcance"},
	}
	for _, c := range casos {
		code, b := d1sPost(t, s.URL, d1sSolveBody(c.in))
		if code != 200 || !strings.Contains(b, c.want) {
			t.Errorf("input %q: esperado 200 con %q, obtenido %d: %.160s", c.in, c.want, code, b)
		}
	}
}

func TestD1ServeStream(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	c, b := d1sPost(t, s.URL, `{"model":"alphazero","stream":true,"messages":[{"role":"user","content":"hola"}]}`)
	if c != 200 || !strings.Contains(b, "chat.completion.chunk") || !strings.Contains(b, "data: [DONE]") {
		t.Errorf("stream malformado: %d: %.200s", c, b)
	}
}

func TestD1ServeModels(t *testing.T) {
	s, _ := newD1sTestServer(t, nil)
	resp, err := http.Get(s.URL + "/v1/models")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"id":"alphazero"`) {
		t.Errorf("models: %d %.200s", resp.StatusCode, b)
	}
}

// Hog sintetico de 128 MiB con limite supervisado de 96 MiB: la guardia debe
// matar al worker, el servidor sigue vivo y no queda residual.
func TestD1ServeHogKilledServerAlive(t *testing.T) {
	s, sv := newD1sTestServer(t, nil)
	sv.testWorkerArgs = []string{sv.exe, "d1hog", "-mb", "128", "-sleep", "10"}
	c, b := d1sPost(t, s.URL, d1sSolveBody("add(va,c0) -> va"))
	if c != 500 || !strings.Contains(b, "memoria") {
		t.Errorf("hog: esperado 500 por memoria, obtenido %d: %.200s", c, b)
	}
	if sv.killed.Load() != 0 {
		t.Errorf("el kill debio venir del guard, no del servidor: killed=%d", sv.killed.Load())
	}
	// servidor vivo tras el fallo del worker
	code, b2 := d1sPost(t, s.URL, d1sSolveBody("hola"))
	if code != 200 || !strings.Contains(b2, "Hola.") {
		t.Errorf("servidor no sano tras hog: %d", code)
	}
}

// Worker lento: el timeout del guard lo termina con 504 y el servidor sigue.
func TestD1ServeSlowWorkerTimeout(t *testing.T) {
	s, sv := newD1sTestServer(t, nil)
	sv.wtimeout = 2
	sv.testWorkerArgs = []string{sv.exe, "d1hog", "-mb", "8", "-sleep", "30"}
	t0 := time.Now()
	c, b := d1sPost(t, s.URL, d1sSolveBody("add(va,c0) -> va"))
	if c != 504 || !strings.Contains(b, "tiempo") {
		t.Errorf("lento: esperado 504, obtenido %d: %.200s", c, b)
	}
	if time.Since(t0) > 10*time.Second {
		t.Errorf("tardo demasiado: %v", time.Since(t0))
	}
	code, _ := d1sPost(t, s.URL, d1sSolveBody("hola"))
	if code != 200 {
		t.Errorf("servidor no sano tras timeout: %d", code)
	}
}

// Cliente que cancela mientras espera en cola: su worker NO arranca.
func TestD1ServeCancelEnColaNoArranca(t *testing.T) {
	s, sv := newD1sTestServer(t, nil)
	sv.wtimeout = 3
	sv.testWorkerArgs = []string{sv.exe, "d1hog", "-mb", "8", "-sleep", "2"} // ocupa el unico slot
	// A: ocupa el worker
	var wgA sync.WaitGroup
	wgA.Add(1)
	go func() {
		defer wgA.Done()
		d1sPost(t, s.URL, d1sSolveBody("add(va,c0) -> va"))
	}()
	time.Sleep(300 * time.Millisecond) // A ya esta dentro
	// B: se encola y cancela a los 200 ms
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "POST", s.URL+"/v1/chat/completions", strings.NewReader(d1sSolveBody("add(va,c0) -> va")))
	req.Header.Set("Content-Type", "application/json")
	go func() { time.Sleep(200 * time.Millisecond); cancel() }()
	http.DefaultClient.Do(req) //nolint: cancelado a proposito
	wgA.Wait()
	time.Sleep(500 * time.Millisecond)
	if n := sv.started.Load(); n != 1 {
		t.Errorf("workers arrancados = %d, esperado 1 (el de la cola cancelada no debio arrancar)", n)
	}
}

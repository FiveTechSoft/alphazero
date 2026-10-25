//go:build !js && linux

// d1serve: servidor HTTP local OpenAI-compatible que expone el motor
// MathZero/AlphaZero como "modelo" para OpenCode, siguiendo el patron del
// repo FiveTechSoft/symbols. Codigo NUEVO, solo rama experimental.
//
// SOLO LINUX VERIFICADO: la supervision usa /proc, setsid y kill de grupo;
// Windows NO probado ni soportado por esta guardia (no se afirma portabilidad).
//
// Rutas deterministas (sin modelo de lenguaje en ninguna):
//  1. saludo/identidad  -> plantilla fija declarada
//  2. '<expr> -> <expr>' en notacion prefija -> WORKER SUBPROCESO por
//     solicitud: `guard.sh <t> <rss> <out> <exe> d1solve ...`. Sin globales
//     compartidos entre solicitudes (cada worker es un proceso nuevo; el SO
//     reclama todo al salir). guard.sh es la supervision ya probada en las
//     campanas A1/G1 (copia byte-identica): lidera su grupo, sondea RSS
//     agregado cada 0.1 s y mata el GRUPO por limite RSS, timeout, señal al
//     supervisor o lider que sale con descendientes; reporta residual y sale
//     99 si queda algo. Cancelacion de cliente o del servidor: TERM al grupo
//     del guard (su trap limpia el grupo del worker), KILL de reserva.
//  3. cualquier otra cosa -> abstencion fija
//
// Limites reales (OPENCODE.md): sin cancelacion dura dentro del motor; el
// freno es entrada acotada + cap + RSS supervisado + timeout + 1 worker.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1serve" {
		if err := RunD1Serve(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "d1serve:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

const d1sModelID = "alphazero"

const (
	d1sMaxBody    = 256 << 10
	d1sMaxQuery   = 4096
	d1sMaxExpr    = 200 // bytes por expresion (conservador; medido: 164 MiB pico con 200B/prof16/cap500)
	d1sMaxDepth   = 16
	d1sMaxCap     = 500
	d1sQueueSlots = 4
	d1sQueueWait  = 10 * time.Second
)

const d1sIdentES = "Soy AlphaZero (MathZero), un motor de algebra local: una politica neuronal entrenada por autojuego guia una busqueda sobre reglas de reescritura de expresiones. No soy un modelo de lenguaje: no converso. Resuelvo problemas de algebra escritos como '<expresion> -> <objetivo>' en notacion prefija (operadores add/mul/neg, constantes cN, variables va-vz). Ejemplo: 'add(va,c0) -> va'."
const d1sGreetES = "Hola. No soy un modelo de lenguaje; soy el motor de algebra AlphaZero (MathZero). Dame un problema, por ejemplo: 'add(va,c0) -> va'."
const d1sIdentEN = "I am AlphaZero (MathZero), a local algebra engine: a neural policy trained by self-play guides a search over expression rewrite rules. I am not a language model: I do not chat. I solve algebra problems written as '<expression> -> <target>' in prefix notation (add/mul/neg, constants cN, variables va-vz). Example: 'add(va,c0) -> va'."
const d1sGreetEN = "Hello. I am not a language model; I am the AlphaZero (MathZero) algebra engine. Give me a problem, for example: 'add(va,c0) -> va'."
const d1sAbstain = "Eso esta fuera de mi alcance. AlphaZero (MathZero) solo resuelve problemas de algebra escritos como '<expresion> -> <objetivo>' en notacion prefija (operadores add/mul/neg, constantes cN, variables va-vz); no soy un modelo de lenguaje y no puedo conversar. / Out of scope: AlphaZero (MathZero) only solves algebra problems written as '<expression> -> <target>' in prefix notation (add/mul/neg, cN, va-vz); it is not a language model."

type d1sServer struct {
	model    string
	capN     int
	rssMB    int
	wtimeout int // segundos, presupuesto del worker (guard)
	guard    string
	exe      string
	worker   chan struct{} // 1 worker en paralelo
	queue    chan struct{}
	started  atomic.Int64 // workers arrancados (para pruebas)
	killed   atomic.Int64 // workers matados por supervision/cancelacion (para pruebas)

	testWorkerArgs []string // SOLO tests: sustituye el comando del worker
}

type d1sChatMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type d1sChatReq struct {
	Model    string       `json:"model"`
	Messages []d1sChatMsg `json:"messages"`
	Stream   bool         `json:"stream"`
}

func d1sText(c json.RawMessage) string {
	var s string
	if json.Unmarshal(c, &s) == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(c, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Type == "text" {
				b.WriteString(p.Text)
			}
		}
		return b.String()
	}
	return ""
}

func d1sNorm(s string) string {
	s = strings.ToLower(s)
	r := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")
	s = r.Replace(s)
	var b []rune
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == ' ' {
			b = append(b, c)
		} else {
			b = append(b, ' ')
		}
	}
	return strings.Join(strings.Fields(string(b)), " ")
}

func d1sHas(norm, trigger string) bool {
	return strings.Contains(" "+norm+" ", " "+trigger+" ")
}

func d1sRoute(text string) string {
	n := d1sNorm(text)
	for _, t := range []string{"quien eres", "que eres", "who are you", "what are you", "que sabes hacer", "what can you do", "presentate"} {
		if d1sHas(n, t) {
			return "identity"
		}
	}
	for _, t := range []string{"hola", "hello", "hi", "hey", "buenos dias", "buenas tardes", "buenas noches", "buenas", "que tal", "como estas", "how are you"} {
		if d1sHas(n, t) {
			return "greeting"
		}
	}
	if strings.Contains(text, "->") {
		return "solve"
	}
	return "abstain"
}

func d1sIsEnglish(text string) bool {
	n := d1sNorm(text)
	for _, t := range []string{"hello", "hi", "hey", "who are you", "what are you", "how are you", "what can you do"} {
		if d1sHas(n, t) {
			return true
		}
	}
	return false
}

func d1sDepth(s string) int {
	d, m := 0, 0
	for _, c := range s {
		switch c {
		case '(':
			d++
			if d > m {
				m = d
			}
		case ')':
			d--
		}
	}
	return m
}

func d1sTemplate(route, text string) string {
	switch route {
	case "identity":
		if d1sIsEnglish(text) {
			return d1sIdentEN
		}
		return d1sIdentES
	case "greeting":
		if d1sIsEnglish(text) {
			return d1sGreetEN
		}
		return d1sGreetES
	}
	return d1sAbstain
}

// d1sRender convierte el informe d1solve en el mensaje visible.
func d1sRender(rep *D1SolveReport) string {
	var b strings.Builder
	if rep.Solved {
		fmt.Fprintf(&b, "Resuelto por el motor MathZero (cap %d nodos; %d pasos, %d nodos explorados):\n%s -> %s\nCamino:\n", rep.Cap, len(rep.Path), rep.Nodes, rep.Start, rep.Target)
		for _, st := range rep.Path {
			fmt.Fprintf(&b, "  %s --[%s]--> %s\n", st.From, st.Rule, st.To)
		}
		b.WriteString("(igualdad modulo la normalizacion asociativa-conmutativa del motor)")
	} else {
		fmt.Fprintf(&b, "El motor MathZero no encontro solucion dentro del presupuesto (%d nodos; explorados %d):\n%s -> %s", rep.Cap, rep.Nodes, rep.Start, rep.Target)
	}
	return b.String()
}

// d1sSolve valida, admite y ejecuta UN worker subproceso supervisado.
// Devuelve (content, errHTTP, errMsg). errHTTP == -1: cliente cancelo en cola;
// no responder.
func (sv *d1sServer) d1sSolve(r *http.Request, text string) (string, int, string) {
	parts := strings.SplitN(text, "->", 2)
	start, target := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	for _, e := range []string{start, target} {
		if len(e) == 0 || len(e) > d1sMaxExpr {
			return "", 400, fmt.Sprintf("expresion vacia o demasiado larga (max %d bytes por lado)", d1sMaxExpr)
		}
		if d1sDepth(e) > d1sMaxDepth {
			return "", 400, fmt.Sprintf("expresion demasiado anidada (max %d niveles)", d1sMaxDepth)
		}
	}
	if _, errS := ParsePrefix(start); errS != nil {
		return fmt.Sprintf("No pude interpretar el problema de algebra. Uso: '<expresion> -> <objetivo>' en notacion prefija (add/mul/neg, cN, va-vz). Detalle: %v", errS), 0, ""
	}
	if _, errT := ParsePrefix(target); errT != nil {
		return fmt.Sprintf("No pude interpretar el problema de algebra. Uso: '<expresion> -> <objetivo>' en notacion prefija (add/mul/neg, cN, va-vz). Detalle: %v", errT), 0, ""
	}
	// Admision: cola explicita; un cliente que cancela en cola NO arranca worker.
	select {
	case sv.queue <- struct{}{}:
		defer func() { <-sv.queue }()
	case <-time.After(d1sQueueWait):
		return "", 503, "servidor ocupado: cola de resolucion llena; reintenta en unos segundos"
	case <-r.Context().Done():
		return "", -1, ""
	}
	select {
	case sv.worker <- struct{}{}:
		defer func() { <-sv.worker }()
	case <-r.Context().Done():
		return "", -1, ""
	}
	return sv.runWorker(r, start, target)
}

func (sv *d1sServer) runWorker(r *http.Request, start, target string) (string, int, string) {
	out, err := os.CreateTemp("", "d1solve-*.out")
	if err != nil {
		return "", 500, "no se pudo crear fichero temporal"
	}
	outPath := out.Name()
	out.Close()
	defer os.Remove(outPath)

	var args []string
	if sv.testWorkerArgs != nil {
		args = append([]string{}, sv.testWorkerArgs...)
	} else {
		args = []string{sv.exe, "d1solve", "-start", start, "-target", target, "-cap", fmt.Sprint(sv.capN)}
		if sv.model != "" {
			args = append(args, "-model", sv.model)
		}
	}
	full := append([]string{fmt.Sprint(sv.wtimeout), fmt.Sprint(sv.rssMB), outPath}, args...)
	cmd := exec.Command("bash", append([]string{sv.guard}, full...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var report bytes.Buffer
	cmd.Stdout = &report
	cmd.Stderr = &report
	if err := cmd.Start(); err != nil {
		return "", 500, "no se pudo arrancar el worker: " + err.Error()
	}
	sv.started.Add(1)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	killGroup := func() {
		sv.killed.Add(1)
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM) // trap del guard: limpia el grupo del worker
		}
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			syscall.Kill(-pgid, syscall.SIGKILL)
			<-done
		}
	}
	select {
	case err := <-done:
		rep := report.String()
		switch {
		case strings.Contains(rep, "status=RSS_LIMIT"):
			return "", 500, fmt.Sprintf("el problema excedio el limite de memoria supervisado (%d MiB); prueba una expresion mas simple o un cap menor", sv.rssMB)
		case strings.Contains(rep, "status=TIMEOUT"):
			return "", 504, fmt.Sprintf("el motor excedio su presupuesto de tiempo (%d s); el worker fue terminado y recolectado", sv.wtimeout)
		case strings.Contains(rep, "status=SAMPLER_FAILED"):
			return "", 500, "la supervision de recursos fallo; solicitud abortada por precaucion"
		case strings.Contains(rep, "residual_procs=0") == false:
			return "", 500, "la supervision detecto procesos residuales; solicitud abortada"
		}
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() != 0 {
			b, _ := os.ReadFile(outPath)
			msg := strings.TrimSpace(string(b))
			if msg == "" {
				msg = fmt.Sprintf("worker termino con codigo %d", exit.ExitCode())
			}
			if len(msg) > 400 {
				msg = msg[:400]
			}
			return fmt.Sprintf("No pude interpretar/ejecutar el problema. Detalle: %s", msg), 0, ""
		}
		b, rerr := os.ReadFile(outPath)
		if rerr != nil {
			return "", 500, "no se pudo leer la salida del worker"
		}
		var d1rep D1SolveReport
		if jerr := json.Unmarshal(b, &d1rep); jerr != nil {
			return "", 500, "salida del worker no interpretable"
		}
		return d1sRender(&d1rep), 0, ""
	case <-r.Context().Done():
		killGroup()
		return "", -1, ""
	case <-time.After(time.Duration(sv.wtimeout+15) * time.Second):
		// Red de seguridad por encima del timeout propio del guard.
		killGroup()
		return "", 504, "el worker excedio la supervision externa; terminado"
	}
}

func d1sJSONEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func d1sError(w http.ResponseWriter, code int, msg string) {
	if code == 503 {
		w.Header().Set("Retry-After", "5")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":{"message":%s,"type":"invalid_request_error","code":null}}`, d1sJSONEscape(msg))
}

func d1sRespond(w http.ResponseWriter, stream bool, content string) {
	id := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()
	if stream {
		chunk1 := fmt.Sprintf(`{"id":%q,"object":"chat.completion.chunk","created":%d,"model":%q,"choices":[{"index":0,"delta":{"role":"assistant","content":%s},"finish_reason":null}]}`, id, created, d1sModelID, d1sJSONEscape(content))
		chunk2 := fmt.Sprintf(`{"id":%q,"object":"chat.completion.chunk","created":%d,"model":%q,"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`, id, created, d1sModelID)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\ndata: %s\n\ndata: [DONE]\n\n", chunk1, chunk2)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id":%q,"object":"chat.completion","created":%d,"model":%q,"choices":[{"index":0,"message":{"role":"assistant","content":%s},"finish_reason":"stop"}],"usage":{"prompt_tokens":0,"completion_tokens":0,"total_tokens":0}}`, id, created, d1sModelID, d1sJSONEscape(content))
}

func (sv *d1sServer) handleCompletions(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, d1sMaxBody)
	var req d1sChatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			d1sError(w, 413, fmt.Sprintf("cuerpo demasiado grande (max %d bytes)", d1sMaxBody))
		} else {
			d1sError(w, 400, "peticion JSON invalida")
		}
		return
	}
	var text string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			text = d1sText(req.Messages[i].Content)
			break
		}
	}
	text = strings.TrimSpace(text)
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		text = strings.TrimSpace(text[1 : len(text)-1]) // OpenCode entrecomilla el mensaje (observado 2026-10-25)
	}
	if len(text) > d1sMaxQuery {
		d1sError(w, 400, fmt.Sprintf("mensaje demasiado largo (max %d bytes)", d1sMaxQuery))
		return
	}
	if text == "" {
		d1sError(w, 400, "sin mensaje de usuario")
		return
	}
	route := d1sRoute(text)
	if route != "solve" {
		d1sRespond(w, req.Stream, d1sTemplate(route, text))
		return
	}
	content, errHTTP, errMsg := sv.d1sSolve(r, text)
	if errHTTP == -1 {
		return // cliente cancelado; nada que responder
	}
	if errHTTP != 0 {
		d1sError(w, errHTTP, errMsg)
		return
	}
	d1sRespond(w, req.Stream, content)
}

func (sv *d1sServer) handleModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"object":"list","data":[{"id":%q,"object":"model","created":0,"owned_by":"alphazero"}]}`, d1sModelID)
}

func RunD1Serve(args []string) error {
	fs := flag.NewFlagSet("d1serve", flag.ContinueOnError)
	port := fs.Int("port", 8100, "puerto local")
	model := fs.String("model", "", "model json (politica entrenada); vacio = MLP nuevo semilla 1")
	capN := fs.Int("cap", 200, "presupuesto de nodos por problema (max 500)")
	rssMB := fs.Int("rssmb", 384, "limite RSS supervisado por worker en MiB (medido: 25 MiB tipico, 164 MiB peor caso admitido)")
	wtimeout := fs.Int("wtimeout", 20, "presupuesto de tiempo del worker en segundos")
	guard := fs.String("guard", "guard.sh", "ruta a guard.sh (copia byte-identica de la supervision probada en campanas A1/G1)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *capN <= 0 {
		return fmt.Errorf("cap debe ser positivo")
	}
	if *capN > d1sMaxCap {
		fmt.Fprintf(os.Stderr, "d1serve: cap %d reducido al maximo %d\n", *capN, d1sMaxCap)
		*capN = d1sMaxCap
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if !filepath.IsAbs(*guard) {
		*guard = filepath.Join(filepath.Dir(exe), *guard)
	}
	for _, f := range []string{*guard, exe} {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("no existe %s: %v", f, err)
		}
	}
	if *model != "" {
		if _, err := os.Stat(*model); err != nil {
			return fmt.Errorf("model: %v", err)
		}
		fmt.Fprintf(os.Stderr, "d1serve: politica %s (sha256 %s)\n", *model, fileSHA(*model))
	} else {
		fmt.Fprintln(os.Stderr, "d1serve: AVISO sin -model, workers con MLP nuevo semilla 1 (sin entrenar)")
	}
	sv := &d1sServer{
		model: *model, capN: *capN, rssMB: *rssMB, wtimeout: *wtimeout,
		guard: *guard, exe: exe,
		worker: make(chan struct{}, 1), queue: make(chan struct{}, d1sQueueSlots),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", sv.handleModels)
	mux.HandleFunc("/v1/chat/completions", sv.handleCompletions)
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	fmt.Fprintf(os.Stderr, "d1serve: escuchando en http://%s/v1 (modelo %q, cap %d, rss %d MiB, worker %d s)\n", addr, d1sModelID, *capN, *rssMB, *wtimeout)
	return srv.ListenAndServe()
}

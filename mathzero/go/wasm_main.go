//go:build js && wasm

// Build WebAssembly del MISMO motor (engine.go, mathring.go, net.go):
// no hay reescritura en JS. La pagina llama a estas funciones y el
// binario ejecuta el mismo codigo Go que la CLI.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"syscall/js"
	"time"
)

//go:embed ttt_model.json
var tttModelJSON []byte

//go:embed model_math.json
var mathModelJSON []byte

//go:embed problems_web.json
var problemsWeb []byte

var (
	tttNet  *TTTNet
	mathNet *MLP
	curTTT  = TTTState{Player: 1}
	rng     = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func tttJSON(extra map[string]any) string {
	m := map[string]any{
		"board":  curTTT.Board,
		"player": curTTT.Player,
		"winner": TTTWinner(curTTT),
	}
	for k, v := range extra {
		m[k] = v
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func jsRet(s string) js.Value { return js.ValueOf(s) }

func main() {
	tttNet = LoadTTTNet(tttModelJSON)
	LemmaLib = NewLibrary()
	mathNet = &MLP{}
	if json.Unmarshal(mathModelJSON, mathNet) != nil {
		mathNet = nil
	}
	fmt.Println("[wasm] motor Go listo; red ttt:", tttNet != nil, "red math:", mathNet != nil)

	js.Global().Set("mzNewTTT", js.FuncOf(func(this js.Value, a []js.Value) any {
		curTTT = TTTState{Player: 1}
		return jsRet(tttJSON(nil))
	}))
	js.Global().Set("mzPlayTTT", js.FuncOf(func(this js.Value, a []js.Value) any {
		i := a[0].Int()
		if TTTWinner(curTTT) == -2 && curTTT.Board[i] == 0 {
			curTTT = TicTacToe{}.Play(curTTT, i).(TTTState)
		}
		return jsRet(tttJSON(nil))
	}))
	js.Global().Set("mzEngineTTT", js.FuncOf(func(this js.Value, a []js.Value) any {
		sims := 50
		if len(a) > 0 {
			sims = a[0].Int()
		}
		if TTTWinner(curTTT) != -2 {
			return jsRet(tttJSON(map[string]any{"move": -1}))
		}
		t0 := time.Now()
		mv := MCTS(curTTT, TicTacToe{}, tttNet, sims, 1.5).(int)
		curTTT = TicTacToe{}.Play(curTTT, mv).(TTTState)
		return jsRet(tttJSON(map[string]any{"move": mv, "ms": time.Since(t0).Milliseconds()}))
	}))
	// Autojuego red contra si misma, con variedad en las 2 primeras
	// jugadas (muestreo por visitas) y argmax despues. Devuelve la
	// partida entera para que la pagina la anime.
	js.Global().Set("mzSelfTTT", js.FuncOf(func(this js.Value, a []js.Value) any {
		sims := 50
		if len(a) > 0 {
			sims = a[0].Int()
		}
		st := TTTState{Player: 1}
		moves := []int{}
		t0 := time.Now()
		for TTTWinner(st) == -2 {
			acts, nvis := MCTSVisits(st, TicTacToe{}, tttNet, sims, 1.5)
			var mv int
			if len(moves) < 2 {
				tot := 0
				for _, n := range nvis {
					tot += n
				}
				if tot == 0 {
					mv = acts[rng.Intn(len(acts))].(int)
				} else {
					r, acc := rng.Intn(tot), 0
					mv = acts[len(acts)-1].(int)
					for j, n := range nvis {
						acc += n
						if r < acc {
							mv = acts[j].(int)
							break
						}
					}
				}
			} else {
				best, bn := 0, -1
				for j, n := range nvis {
					if n > bn {
						bn, best = n, j
					}
				}
				mv = acts[best].(int)
			}
			st = TicTacToe{}.Play(st, mv).(TTTState)
			moves = append(moves, mv)
		}
		b, _ := json.Marshal(map[string]any{
			"moves": moves, "winner": TTTWinner(st), "ms": time.Since(t0).Milliseconds(),
		})
		return jsRet(string(b))
	}))
	js.Global().Set("mzMathList", js.FuncOf(func(this js.Value, a []js.Value) any {
		type P struct {
			Level string `json:"level"`
			Idx   int    `json:"idx"`
			A     string `json:"a"`
			B     string `json:"b"`
		}
		var raw map[string][]map[string]json.RawMessage
		json.Unmarshal(problemsWeb, &raw)
		out := []P{}
		for _, lv := range []string{"3", "4", "5"} {
			for i, pr := range raw[lv] {
				out = append(out, P{lv, i, Render(parseExpr(pr["a"])), Render(parseExpr(pr["b"]))})
			}
		}
		b, _ := json.Marshal(out)
		return jsRet(string(b))
	}))
	// Resuelve con la red entrenada (SolvePath) y devuelve el camino.
	js.Global().Set("mzMathSolve", js.FuncOf(func(this js.Value, a []js.Value) any {
		lv := a[0].String()
		idx := a[1].Int()
		cap := 2000
		if len(a) > 2 {
			cap = a[2].Int()
		}
		probs := LoadProblemsBytes(problemsWeb, lv)
		if idx < 0 || idx >= len(probs) {
			return jsRet(`{"ok":false}`)
		}
		p := probs[idx]
		t0 := time.Now()
		ok, nodes, path := SolvePath(mathNet, p.Cur, p.Target, cap)
		type S struct {
			From string `json:"from"`
			To   string `json:"to"`
			Rule string `json:"rule"`
		}
		steps := []S{}
		for _, st := range path {
			steps = append(steps, S{Render(st.From), Render(st.To), st.Rule})
		}
		b, _ := json.Marshal(map[string]any{
			"ok": ok, "nodes": nodes, "ms": time.Since(t0).Milliseconds(),
			"a": Render(p.Cur), "b": Render(p.Target), "steps": steps,
			"lemas": len(LemmaLib.Lemmas),
		})
		return jsRet(string(b))
	}))
	select {}
}

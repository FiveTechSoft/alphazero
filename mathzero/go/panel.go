// Panel de capacidades (h18, base): un comando que reune en una sola tabla
// las metricas ya medidas en el ciego y las deja en JSON versionable.
//
//	mathzero panel [cap] [modelo] [-json] [-no-ttt] [-nosave] [-datadir DIR]
//
// Cubre: (1) % por nivel k2-k8 ciego, (2) nodos/tiempo/longitud de prueba,
// (3) k2-k5 vs k6-k8 y otro dominio (tres en raya contra minimax exacto),
// (6) estado de la biblioteca de lemas. Quedan fuera Elo entre versiones (5)
// y la referencia externa LLM (7).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type panelNivel struct {
	Nivel     int     `json:"nivel"`
	Problemas int     `json:"problemas"`
	Resueltos int     `json:"resueltos"`
	NodosMed  float64 `json:"nodos_med"`
	PruebaMed float64 `json:"prueba_med"`
	TiempoMs  int64   `json:"tiempo_ms"`
}

type panelTTT struct {
	Partidas int `json:"partidas"`
	Sims     int `json:"sims"`
	Gana     int `json:"gana"`
	Tablas   int `json:"tablas"`
	Pierde   int `json:"pierde"`
}

type panelResult struct {
	Fecha       string       `json:"fecha"`
	Modelo      string       `json:"modelo"`
	Nivel       int          `json:"nivel"`
	Ronda       int          `json:"ronda"`
	Cap         int          `json:"cap"`
	Total       int          `json:"total"`
	Problemas   int          `json:"problemas"`
	K25         int          `json:"k2_5"`
	K68         int          `json:"k6_8"`
	NodosMed    float64      `json:"nodos_med"`
	PruebaMed   float64      `json:"prueba_med"`
	TiempoMs    int64        `json:"tiempo_ms"`
	PorNivel    []panelNivel `json:"por_nivel"`
	LemasTotal  int          `json:"lemas_total"`
	LemasUsados int          `json:"lemas_usados"`
	LemasTop8   int          `json:"lemas_top8"`
	TTT         *panelTTT    `json:"ttt,omitempty"`
}

func evalPanel(net *MLP, lv map[int][]MState, cap int) panelResult {
	var r panelResult
	start := time.Now()
	r.Problemas = 0
	nodosTotal, resueltosTotal, pruebasTotal := 0, 0, 0
	for k := 2; k <= 8; k++ {
		if len(lv[k]) == 0 {
			continue
		}
		kStart := time.Now()
		var niv panelNivel
		niv.Nivel = k
		niv.Problemas = len(lv[k])
		nodos := 0
		for _, p := range lv[k] {
			ok, n, sm := GuidedSolve(net, p.Cur, p.Target, cap)
			nodos += n
			if ok {
				niv.Resueltos++
				niv.PruebaMed += float64(len(sm))
			}
		}
		if niv.Resueltos > 0 {
			niv.PruebaMed /= float64(niv.Resueltos)
		}
		niv.NodosMed = float64(nodos) / float64(niv.Problemas)
		niv.TiempoMs = time.Since(kStart).Milliseconds()
		r.PorNivel = append(r.PorNivel, niv)
		r.Problemas += niv.Problemas
		resueltosTotal += niv.Resueltos
		nodosTotal += nodos
		pruebasTotal += int(niv.PruebaMed * float64(niv.Resueltos))
		if k <= 5 {
			r.K25 += niv.Resueltos
		} else {
			r.K68 += niv.Resueltos
		}
	}
	r.Total = resueltosTotal
	if r.Problemas > 0 {
		r.NodosMed = float64(nodosTotal) / float64(r.Problemas)
	}
	if resueltosTotal > 0 {
		r.PruebaMed = float64(pruebasTotal) / float64(resueltosTotal)
	}
	r.TiempoMs = time.Since(start).Milliseconds()
	return r
}

// tttMinimax devuelve el resultado optimo desde el punto de vista de X
// (+1 gana X, -1 gana O, 0 tablas).
func tttMinimax(st TTTState, memo map[TTTState]int) int {
	if v, ok := memo[st]; ok {
		return v
	}
	if w := TTTWinner(st); w != -2 {
		memo[st] = w
		return w
	}
	best := 0
	first := true
	for i, v := range st.Board {
		if v != 0 {
			continue
		}
		ns := st
		ns.Board[i] = st.Player
		ns.Player = -st.Player
		score := tttMinimax(ns, memo)
		if first {
			best, first = score, false
		} else if st.Player == 1 && score > best {
			best = score
		} else if st.Player == -1 && score < best {
			best = score
		}
	}
	memo[st] = best
	return best
}

func tttPerfectMove(st TTTState) int {
	memo := map[TTTState]int{}
	bestMove := -1
	bestScore := 0
	for i, v := range st.Board {
		if v != 0 {
			continue
		}
		ns := st
		ns.Board[i] = st.Player
		ns.Player = -st.Player
		score := tttMinimax(ns, memo)
		if bestMove < 0 || (st.Player == 1 && score > bestScore) || (st.Player == -1 && score < bestScore) {
			bestMove, bestScore = i, score
		}
	}
	return bestMove
}

// evalTTT juega partidas del modelo contra minimax exacto, alternando color.
func evalTTT(net Net, sims, games int) panelTTT {
	out := panelTTT{Partidas: games, Sims: sims}
	g := TicTacToe{}
	for i := 0; i < games; i++ {
		netPlayer := int8(1)
		if i%2 == 1 {
			netPlayer = -1
		}
		st := TTTState{Player: 1}
		for {
			if w := TTTWinner(st); w != -2 {
				if w == 0 {
					out.Tablas++
				} else if int8(w) == netPlayer {
					out.Gana++
				} else {
					out.Pierde++
				}
				break
			}
			var a any
			if st.Player == netPlayer {
				a = MCTS(st, g, net, sims, 1.5)
			} else {
				a = tttPerfectMove(st)
			}
			if a == nil {
				out.Tablas++
				break
			}
			st = g.Play(st, a).(TTTState)
		}
	}
	return out
}

func PanelMain(args []string) {
	var ok bool
	if args, ok = parseDataFlags(args); !ok {
		return
	}
	wantJSON, runTTT := false, true
	var pos []string
	for _, a := range args {
		switch a {
		case "-json":
			wantJSON = true
		case "-no-ttt":
			runTTT = false
		default:
			pos = append(pos, a)
		}
	}
	if len(pos) > 2 {
		fmt.Fprintln(os.Stderr, "uso: mathzero panel [cap] [modelo] [-json] [-no-ttt] [-nosave] [-datadir DIR]")
		return
	}
	cap, model := 1000, dataPath("model_math.json")
	if len(pos) > 0 {
		if cap, ok = parseIntArg("cap", pos[0], 1); !ok {
			return
		}
	}
	if len(pos) > 1 {
		model = pos[1]
	}
	net := LoadMLP(model)
	if net == nil {
		fmt.Fprintln(os.Stderr, "[panel] no se pudo cargar el modelo:", model)
		return
	}
	DirFix = true
	LemmaLib = nil
	lv := map[int][]MState{}
	for k := 2; k <= 8; k++ {
		lv[k] = LoadProblems(dataPath("blind.json"), fmt.Sprint(k))
	}
	lib := LoadLibrary(dataPath("lemas.json"))
	res := evalPanel(net, lv, cap)
	res.Fecha = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	res.Modelo = model
	res.Nivel, res.Ronda, res.Cap = net.Level, net.Round, cap
	res.LemasTotal = len(lib.Lemmas)
	res.LemasUsados = len(lib.UsoMap())
	res.LemasTop8 = len(lib.Pruned(lib.UsoMap(), 8).Lemmas)
	if runTTT {
		if data := readDataFile(dataPath("ttt_model.json")); data != nil {
			if tn := LoadTTTNet(data); tn != nil {
				t := evalTTT(tn, 50, 50)
				res.TTT = &t
			}
		}
	}

	fmt.Printf("[panel] modelo=%s nivel=%d ronda=%d cap=%d fecha=%s\n",
		res.Modelo, res.Nivel, res.Ronda, res.Cap, res.Fecha)
	for _, n := range res.PorNivel {
		fmt.Printf("[panel] k%d: %d/%d nodos_med=%.0f prueba_med=%.1f tiempo=%dms\n",
			n.Nivel, n.Resueltos, n.Problemas, n.NodosMed, n.PruebaMed, n.TiempoMs)
	}
	fmt.Printf("[panel] total=%d/%d k2-5=%d k6-8=%d nodos_med=%.0f prueba_med=%.1f tiempo=%dms\n",
		res.Total, res.Problemas, res.K25, res.K68, res.NodosMed, res.PruebaMed, res.TiempoMs)
	fmt.Printf("[panel] lemas: total=%d usados=%d top8=%d\n", res.LemasTotal, res.LemasUsados, res.LemasTop8)
	if res.TTT != nil {
		fmt.Printf("[panel] tres-en-raya vs minimax: %d partidas, gana=%d tablas=%d pierde=%d (sims=%d)\n",
			res.TTT.Partidas, res.TTT.Gana, res.TTT.Tablas, res.TTT.Pierde, res.TTT.Sims)
	}
	if wantJSON {
		data, _ := json.MarshalIndent(res, "", "  ")
		name := dataPath("panel.json")
		if err := saveJSON(name, data); err != nil {
			fmt.Fprintln(os.Stderr, "[panel] error escribiendo", name, ":", err)
			return
		}
		if !NoSave {
			fmt.Println("[panel] JSON escrito en", name)
		}
	}
}

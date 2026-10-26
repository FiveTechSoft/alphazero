//go:build !js

// MathZero - binario CLI. El nucleo (Game, MCTS, TicTacToe) vive en
// engine.go, compartido con el build WebAssembly.
package main

import (
	"fmt"
	"os"
	"time"
)

func usage() {
	fmt.Print(`MathZero - motor de autojuego con juez exacto

Uso:
  mathzero                          prueba rapida (bench MCTS + juez + BFS)
  mathzero help                     muestra esta ayuda
  mathzero benchttt                 bench del tres en raya (sin red)
  mathzero learn [-nosave]          entrenamiento con curriculo (12 rondas)
  mathzero cycle [n] [per] [cap]    ciclo de automejora
  mathzero curve [n] [per] [cap]    curva sobre el conjunto ciego
  mathzero evalblind [cap] [modelo] [lam] evaluacion solo-lectura en el ciego (+lam: feedback freqrule)
  mathzero panel [cap] [modelo]     panel de capacidades (ciego + TTT)
  mathzero maze [n] [sims] [lado]   laberinto MCTS vs BFS (h9)
  mathzero discover [n] [cap] [semilla]  busca lemas nuevos (h15)
  mathzero serve [addr]             API HTTP local (Q8)
  mathzero ab <exp> <semilla> ...   A/B con el mismo ciego
  mathzero ab <exp> <semilla> [rondas] [per] [cap] [-json]
  mathzero evalfail [datos] [modelo] [capNet] [capOr] [kmin] [kmax]
  mathzero judgesv                  exporta/verifica vectores del juez
  mathzero genblind                 regenera blind.json (semilla 12345)

Experimentos de "ab": e2, e4, e8, e9, e10, e11, e12, e13.
Los comandos cycle/curve escriben model_math.json y lemas.json.
Flags comunes: -nosave (no escribe nada) y -datadir DIR (lee/escribe ahi).
`)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			usage()
			return
		}
	}
	if len(os.Args) > 1 && os.Args[1] == "learn" {
		LearnMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "benchttt" {
		g := TicTacToe{}
		games, sims := 200, 100
		totalSims := 0
		start := time.Now()
		for i := 0; i < games; i++ {
			s := any(TTTState{Player: 1})
			for {
				done, _ := g.Terminal(s)
				if done {
					break
				}
				a := MCTS(s, g, nil, sims, 1.5)
				totalSims += sims
				s = g.Play(s, a)
			}
		}
		el := time.Since(start).Seconds()
		fmt.Printf("[benchttt Go] %d partidas x %d sims = %d sims en %.3fs -> %.0f sims/s (%.1f partidas/s)\n",
			games, sims, totalSims, el, float64(totalSims)/el, float64(games)/el)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "judgesv" {
		JudgeSV()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "ab" {
		ABMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "evalfail" {
		EvalFail(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "genblind" {
		GenBlind("blind.json", 100, 12345)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "evalblind" {
		EvalBlindMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		ServeMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "discover" {
		DiscoverMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "maze" {
		MazeMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "panel" {
		PanelMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "curve" {
		CurveMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "cycle" {
		CycleMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "subcomando desconocido: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	// 1) TicTacToe: bench de velocidad del motor (sin red)
	g := TicTacToe{}
	games, sims := 100, 50
	totalSims := 0
	start := time.Now()
	for i := 0; i < games; i++ {
		s := any(TTTState{Player: 1})
		for {
			done, _ := g.Terminal(s)
			if done {
				break
			}
			a := MCTS(s, g, nil, sims, 1.5)
			totalSims += sims
			s = g.Play(s, a)
		}
	}
	el := time.Since(start).Seconds()
	fmt.Printf("[TicTacToe] %d partidas x %d sims = %d sims en %.3fs -> %.0f sims/s\n",
		games, sims, totalSims, el, float64(totalSims)/el)

	// 2) MathRing: chequeo del juez exacto + baseline BFS (mismo conjunto que Python)
	mg := MathRing{}
	CheckRules(mg, 200)
	for _, cfg := range []struct {
		k   string
		cap int
	}{{"4", 2000}, {"6", 4000}} {
		probs := LoadProblems("problems.json", cfg.k)
		if len(probs) == 0 {
			continue
		}
		ok, nodes := 0, 0
		for _, p := range probs {
			solved, n := BFSSolve(mg, p, cfg.cap)
			if solved {
				ok++
			}
			nodes += n
		}
		fmt.Printf("[MathRing] baseline BFS k=%s cap=%d: resueltos %d/%d, media nodos %d\n",
			cfg.k, cfg.cap, ok, len(probs), nodes/len(probs))
	}
}

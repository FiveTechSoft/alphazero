//go:build !js

// MathZero - binario CLI. El nucleo (Game, MCTS, TicTacToe) vive en
// engine.go, compartido con el build WebAssembly.
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "learn" {
		Curriculum(12, 150, 2000, "model_math.json", 1)
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
	if len(os.Args) > 1 && os.Args[1] == "curve" {
		CurveMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "cycle" {
		CycleMain(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "geometry" {
		GeomMain(os.Args[2:])
		return
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

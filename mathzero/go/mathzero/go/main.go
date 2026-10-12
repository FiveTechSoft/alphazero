// MathZero - nucleo compilado del motor AlphaZero generico.
// Interfaz Game: estados, acciones legales, transicion y test terminal
// con recompensa. Implementaciones sobre el MISMO motor MCTS:
// TicTacToe (tres en raya) y MathRing (demostracion de identidades
// algebraicas con juez exacto por forma canonica de polinomio).
package main

import (
	"fmt"
	"math"
	"time"
)

// ---------------- interfaz generica ----------------

type Game interface {
	Legal(s any) []any               // acciones legales desde s
	Play(s any, a any) any           // transicion
	Terminal(s any) (bool, float64)  // fin?, valor para quien busca
}

// ---------------- motor MCTS generico (PUCT) ----------------
// Red enchufable: si es nil, prior uniforme y valor neutro (modo bench).

type Net interface {
	Predict(s any, acts []any) (priors []float64, value float64)
}

type Node struct {
	N        []int
	W        []float64
	P        []float64
	Acts     []any
	Children map[int]*Node
	Expanded bool
}

func expand(n *Node, s any, g Game, net Net) float64 {
	n.Acts = g.Legal(s)
	n.N = make([]int, len(n.Acts))
	n.W = make([]float64, len(n.Acts))
	n.P = make([]float64, len(n.Acts))
	n.Children = make(map[int]*Node, len(n.Acts))
	if net == nil {
		for i := range n.P {
			n.P[i] = 1.0 / float64(len(n.P))
		}
		return 0.0
	}
	p, v := net.Predict(s, n.Acts)
	copy(n.P, p)
	return v
}

// MCTS ejecuta sims simulaciones PUCT y devuelve la accion mas visitada.
func MCTS(s any, g Game, net Net, sims int, cPuct float64) any {
	root := &Node{}
	expand(root, s, g, net)
	for i := 0; i < sims; i++ {
		cur := s
		node := root
		type step struct {
			nd *Node
			j  int
		}
		var path []step
		done, v := g.Terminal(cur)
		for !done && node.Expanded {
			total := 0
			for _, x := range node.N {
				total += x
			}
			best, bestScore := -1, -1e18
			for j := range node.P {
				q := 0.0
				if node.N[j] > 0 {
					q = node.W[j] / float64(node.N[j])
				}
				u := cPuct * node.P[j] * math.Sqrt(float64(total+1)) / (1 + float64(node.N[j]))
				if q+u > bestScore {
					bestScore, best = q+u, j
				}
			}
			path = append(path, step{node, best})
			cur = g.Play(cur, node.Acts[best])
			child, ok := node.Children[best]
			if !ok {
				child = &Node{}
				node.Children[best] = child
			}
			node = child
			done, v = g.Terminal(cur)
		}
		if !done {
			v = expand(node, cur, g, net)
		}
		for k := len(path) - 1; k >= 0; k-- {
			v = -v
			path[k].nd.N[path[k].j]++
			path[k].nd.W[path[k].j] += v
		}
	}
	best, bestN := 0, -1
	for j, n := range root.N {
		if n > bestN {
			bestN, best = n, j
		}
	}
	return root.Acts[best]
}

// ---------------- TicTacToe sobre la interfaz ----------------

type TTTState struct {
	Board  [9]int8
	Player int8
}

type TicTacToe struct{}

var winsTTT = [][3]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, {0, 4, 8}, {2, 4, 6}}

func (TicTacToe) Legal(s any) []any {
	st := s.(TTTState)
	var out []any
	for i, v := range st.Board {
		if v == 0 {
			out = append(out, i)
		}
	}
	return out
}

func (TicTacToe) Play(s any, a any) any {
	st := s.(TTTState)
	st.Board[a.(int)] = st.Player
	st.Player = -st.Player
	return st
}

func (TicTacToe) Terminal(s any) (bool, float64) {
	st := s.(TTTState)
	for _, w := range winsTTT {
		sum := st.Board[w[0]] + st.Board[w[1]] + st.Board[w[2]]
		if sum == 3 || sum == -3 {
			return true, -1.0
		}
	}
	for _, v := range st.Board {
		if v == 0 {
			return false, 0
		}
	}
	return true, 0.0
}

// ---------------- main: test de las DOS games en el mismo ejecutable ----------------

func main() {
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
	}{{ "4", 2000 }, { "6", 4000 }} {
		probs := LoadProblems("problems.json", cfg.k)
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

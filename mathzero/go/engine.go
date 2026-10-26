// Nucleo generico: interfaz Game + motor MCTS (PUCT). Compartido por el
// binario CLI (!js) y por el build WebAssembly (js), mismo codigo.
package main

import "math"

// ---------------- interfaz generica ----------------

type Game interface {
	Legal(s any) []any              // acciones legales desde s
	Play(s any, a any) any          // transicion
	Terminal(s any) (bool, float64) // fin?, valor para quien busca
}

// Red enchufable: si es nil, prior uniforme y valor neutro (modo bench).
type Net interface {
	Predict(s any, acts []any) (priors []float64, value float64)
}

// SinglePlayer marca juegos de un solo agente: el valor terminal es para el
// mismo jugador en todos los plies, asi que el backup NO cambia de signo.
type SinglePlayer interface {
	SinglePlayer() bool
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
	n.Expanded = true
	if net == nil {
		setUniformPriors(n.P)
		return 0.0
	}
	p, v := net.Predict(s, n.Acts)
	if !setPriors(n.P, p) {
		setUniformPriors(n.P)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	return v
}

func setUniformPriors(p []float64) {
	if len(p) == 0 {
		return
	}
	u := 1.0 / float64(len(p))
	for i := range p {
		p[i] = u
	}
}

// setPriors copia y normaliza las priors de la red. Si la red devuelve
// un vector invalido (longitud, valores negativos o no finitos, suma no
// positiva), el llamador cae a prior uniforme.
func setPriors(dst, src []float64) bool {
	if len(dst) != len(src) || len(dst) == 0 {
		return false
	}
	sum := 0.0
	for _, x := range src {
		if math.IsNaN(x) || math.IsInf(x, 0) || x < 0 {
			return false
		}
		sum += x
	}
	if sum <= 0 {
		return false
	}
	for i, x := range src {
		dst[i] = x / sum
	}
	return true
}

// MCTSVisits ejecuta sims simulaciones PUCT y devuelve las acciones de la
// raiz con sus visitas (para argmax o muestreo con temperatura).
func MCTSVisits(s any, g Game, net Net, sims int, cPuct float64) ([]any, []int) {
	if done, _ := g.Terminal(s); done {
		return nil, nil
	}
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
			if best < 0 {
				break
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
		_, single := g.(SinglePlayer)
		for k := len(path) - 1; k >= 0; k-- {
			if !single {
				v = -v
			}
			path[k].nd.N[path[k].j]++
			path[k].nd.W[path[k].j] += v
		}
	}
	return root.Acts, root.N
}

// MCTS devuelve la accion mas visitada.
func MCTS(s any, g Game, net Net, sims int, cPuct float64) any {
	acts, n := MCTSVisits(s, g, net, sims, cPuct)
	if len(acts) == 0 {
		return nil
	}
	best, bestN := 0, -1
	for j, x := range n {
		if x > bestN {
			bestN, best = x, j
		}
	}
	return acts[best]
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

// TTTWinner: 1 gana X, -1 gana O, 0 tablas, -2 sigue el juego.
func TTTWinner(st TTTState) int {
	for _, w := range winsTTT {
		sum := st.Board[w[0]] + st.Board[w[1]] + st.Board[w[2]]
		if sum == 3 {
			return 1
		}
		if sum == -3 {
			return -1
		}
	}
	for _, v := range st.Board {
		if v == 0 {
			return -2
		}
	}
	return 0
}

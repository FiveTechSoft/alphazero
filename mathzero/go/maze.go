// Laberinto (hito 9): un Game mas sobre el mismo motor MCTS/PUCT, medido
// contra BFS. Los laberintos se generan con semilla fija (DFS aleatorio),
// de modo que el comando es repetible.
//
//	mathzero maze [laberintos] [sims] [lado]
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Maze struct {
	W, H, Goal int
	Walls      []bool
	MaxSteps   int
}

type MazeState struct {
	Pos   int
	Steps int
	Seen  [2]uint64
}

func seenBit(pos int) (int, uint64) { return pos / 64, uint64(1) << uint(pos%64) }

func (st MazeState) seen(pos int) bool {
	w, mask := seenBit(pos)
	return st.Seen[w]&mask != 0
}

func (st MazeState) mark(pos int) MazeState {
	w, mask := seenBit(pos)
	st.Seen[w] |= mask
	return st
}

func (m Maze) Initial() MazeState {
	st := MazeState{Pos: m.Start()}
	st.Seen = [2]uint64{}
	return st.mark(st.Pos)
}

func (Maze) SinglePlayer() bool { return true }

func (m Maze) idx(x, y int) int { return y*m.W + x }

func (m Maze) carve(x, y int) { m.Walls[m.idx(x, y)] = false }

// NewMaze genera un laberinto (lado impar) con DFS aleatorio.
func NewMaze(lado int, seed int64) Maze {
	if lado < 5 {
		lado = 5
	}
	if lado%2 == 0 {
		lado++
	}
	m := Maze{W: lado, H: lado}
	m.Walls = make([]bool, lado*lado)
	for i := range m.Walls {
		m.Walls[i] = true
	}
	m.MaxSteps = lado * lado * 2
	m.Goal = m.idx(lado-2, lado-2)
	r := rand.New(rand.NewSource(seed))
	stack := [][2]int{{1, 1}}
	m.carve(1, 1)
	dirs := [][2]int{{0, -2}, {0, 2}, {-2, 0}, {2, 0}}
	for len(stack) > 0 {
		c := stack[len(stack)-1]
		var opts [][2]int
		for _, d := range dirs {
			nx, ny := c[0]+d[0], c[1]+d[1]
			if nx > 0 && nx < lado-1 && ny > 0 && ny < lado-1 && m.Walls[m.idx(nx, ny)] {
				opts = append(opts, [2]int{nx, ny})
			}
		}
		if len(opts) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}
		n := opts[r.Intn(len(opts))]
		m.carve((c[0]+n[0])/2, (c[1]+n[1])/2)
		m.carve(n[0], n[1])
		stack = append(stack, n)
	}
	return m
}

func (m Maze) xy(pos int) (int, int) { return pos % m.W, pos / m.W }

func (m Maze) Start() int { return m.idx(1, 1) }

func (m Maze) Legal(s any) []any {
	st := s.(MazeState)
	x, y := m.xy(st.Pos)
	var out []any
	for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nx, ny := x+d[0], y+d[1]
		if nx < 0 || nx >= m.W || ny < 0 || ny >= m.H {
			continue
		}
		pos := m.idx(nx, ny)
		if !m.Walls[pos] && !st.seen(pos) {
			out = append(out, pos)
		}
	}
	return out
}

func (m Maze) Play(s any, a any) any {
	st := s.(MazeState)
	st.Pos = a.(int)
	st.Steps++
	return st.mark(st.Pos)
}

// Terminal en un juego de un solo jugador: +1 si llego a la meta, -1 si
// agoto el presupuesto de pasos (MCTS no cambia el signo en estos juegos).
func (m Maze) Terminal(s any) (bool, float64) {
	st := s.(MazeState)
	if st.Pos == m.Goal {
		return true, 1
	}
	if st.Steps >= m.MaxSteps || len(m.Legal(st)) == 0 {
		return true, -1
	}
	return false, 0
}

func manhattan(m Maze, pos int) int {
	x, y := m.xy(pos)
	gx, gy := m.xy(m.Goal)
	d := x - gx
	if d < 0 {
		d = -d
	}
	e := y - gy
	if e < 0 {
		e = -e
	}
	return d + e
}

// MazeNet: prior proporcional a la cercania Manhattan; sin valor aprendido.
type MazeNet struct{ M Maze }

func (n MazeNet) Predict(s any, acts []any) ([]float64, float64) {
	p := make([]float64, len(acts))
	for i, a := range acts {
		p[i] = 1.0 / float64(1+manhattan(n.M, a.(int)))
	}
	return p, 0
}

func mazeBFS(m Maze) (bool, int, int) {
	start := m.Initial()
	goal := m.Goal
	seen := map[int]bool{start.Pos: true}
	queue := []MazeState{start}
	nodes := 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		nodes++
		if cur.Pos == goal {
			return true, nodes, cur.Steps
		}
		for _, a := range m.Legal(cur) {
			nx := m.Play(cur, a).(MazeState)
			if !seen[nx.Pos] {
				seen[nx.Pos] = true
				queue = append(queue, nx)
			}
		}
	}
	return false, nodes, 0
}

func mazeMCTS(m Maze, sims int) (bool, int) {
	st := any(m.Initial())
	net := MazeNet{M: m}
	for step := 0; step < m.MaxSteps; step++ {
		if done, _ := m.Terminal(st); done {
			return st.(MazeState).Pos == m.Goal, step
		}
		a := MCTS(st, m, net, sims, 1.5)
		if a == nil {
			return false, step
		}
		st = m.Play(st, a)
	}
	return false, m.MaxSteps
}

func MazeMain(args []string) {
	rest, ok := parseDataFlags(args)
	if !ok {
		return
	}
	if len(rest) > 3 {
		fmt.Fprintln(os.Stderr, "uso: mathzero maze [laberintos] [sims] [lado]")
		return
	}
	n, sims, lado := 5, 200, 9
	if len(rest) > 0 {
		if n, ok = parseIntArg("laberintos", rest[0], 1); !ok {
			return
		}
	}
	if len(rest) > 1 {
		if sims, ok = parseIntArg("sims", rest[1], 1); !ok {
			return
		}
	}
	if len(rest) > 2 {
		if lado, ok = parseIntArg("lado", rest[2], 5); !ok {
			return
		}
	}
	bfsOK, bfsNodes, bfsSteps := 0, 0, 0
	mctsOK, mctsSteps := 0, 0
	var elBFS, elMCTS time.Duration
	for i := 0; i < n; i++ {
		m := NewMaze(lado, int64(i+1))
		t := time.Now()
		okB, nodes, steps := mazeBFS(m)
		elBFS += time.Since(t)
		if okB {
			bfsOK++
			bfsNodes += nodes
			bfsSteps += steps
		}
		t = time.Now()
		okM, ms := mazeMCTS(m, sims)
		elMCTS += time.Since(t)
		if okM {
			mctsOK++
			mctsSteps += ms
		}
	}
	fmt.Printf("[maze] %d laberintos %dx%d, BFS: %d/%d nodos_med=%d pasos_med=%d tiempo=%s\n",
		n, lado, lado, bfsOK, n, bfsNodes/maxInt(bfsOK, 1), bfsSteps/maxInt(bfsOK, 1), elBFS)
	fmt.Printf("[maze] MCTS (%d sims, prior Manhattan): %d/%d pasos_med=%d tiempo=%s\n",
		sims, mctsOK, n, mctsSteps/maxInt(mctsOK, 1), elMCTS)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

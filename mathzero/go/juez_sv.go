// Exporta vectores de prueba del juez (forma canonica + BFS) para
// verificar el port a Rust: 0 discrepancias exigidas. Tambien mide la
// velocidad del juez Go: juicios/s (canonEq) y nodos/s (BFS).
//
//	./mathzero judgesv
//
// Escribe juez_vectores.txt: una linea por caso:
//
//	<exprA> ; <exprB> ; <canonEq 0/1> ; <bfs_solved 0/1> ; <bfs_nodos>
//
// Expresiones en formato prefijo: c5, va, neg(e), add(e,e), mul(e,e).
package main

import (
	"fmt"
	"os"
	"time"
)

func RenderPrefix(e *Expr) string {
	switch e.Op {
	case "const":
		return fmt.Sprintf("c%d", e.N)
	case "var":
		return "v" + e.S
	case "neg":
		return "neg(" + RenderPrefix(e.A) + ")"
	}
	return e.Op + "(" + RenderPrefix(e.A) + "," + RenderPrefix(e.B) + ")"
}

func JudgeSV() {
	type vec struct {
		a, b   *Expr
		eq     bool
		solved bool
		nodes  int
	}
	var vs []vec
	// 1) casos limite hechos a mano
	E := func(a, b *Expr) { vs = append(vs, vec{a: a, b: b}) }
	E(C(0), C(0))
	E(C(0), C(1))
	E(C(-3), C(-3))
	E(V("a"), V("a"))
	E(V("a"), V("b"))
	E(EAdd(V("a"), C(0)), V("a"))
	E(EAdd(C(0), V("a")), V("a"))
	E(EMul(V("a"), C(1)), V("a"))
	E(EMul(C(0), V("a")), C(0))
	E(ENeg(ENeg(V("a"))), V("a"))
	E(EAdd(V("a"), V("b")), EAdd(V("b"), V("a")))
	E(EMul(V("a"), V("b")), EMul(V("b"), V("a")))
	E(EAdd(V("a"), ENeg(V("a"))), C(0))
	E(EMul(C(2), V("a")), EAdd(V("a"), V("a")))
	E(EMul(V("a"), EAdd(V("b"), V("c"))), EAdd(EMul(V("a"), V("b")), EMul(V("a"), V("c"))))
	E(EMul(EAdd(V("a"), V("b")), EAdd(V("a"), ENeg(V("b")))), EAdd(EMul(V("a"), V("a")), ENeg(EMul(V("b"), V("b")))))
	E(EAdd(EMul(C(2), V("a")), EMul(C(3), V("a"))), EMul(C(5), V("a")))
	E(EMul(EMul(V("a"), V("a")), V("a")), EMul(V("a"), EMul(V("a"), V("a"))))
	E(EAdd(C(1), EAdd(C(2), C(3))), C(6))
	E(EMul(C(2), EMul(C(3), C(-1))), C(-6))
	// 2) conjunto fijo levels.json (30 por nivel)
	for k := 2; k <= 8; k++ {
		for _, p := range LoadProblems("levels.json", fmt.Sprint(k)) {
			vs = append(vs, vec{a: p.Cur, b: p.Target})
		}
	}
	// 3) rellena veredictos con el juez Go
	for i := range vs {
		vs[i].eq = canonEq(Canon(vs[i].a), Canon(vs[i].b))
		vs[i].solved, vs[i].nodes = BFSSolve(MathRing{}, MState{Cur: vs[i].a, Target: vs[i].b, MaxSteps: 64}, 2000)
	}
	f, _ := os.Create("juez_vectores.txt")
	defer f.Close()
	for _, v := range vs {
		fmt.Fprintf(f, "%s ; %s ; %d ; %d ; %d\n", RenderPrefix(v.a), RenderPrefix(v.b), b2i(v.eq), b2i(v.solved), v.nodes)
	}
	fmt.Printf("[judgesv] %d vectores en juez_vectores.txt\n", len(vs))
	// 4) velocidad del juez Go: juicios canon/s
	reps := 200
	t0 := time.Now()
	n := 0
	for r := 0; r < reps; r++ {
		for _, v := range vs {
			_ = canonEq(Canon(v.a), Canon(v.b))
			n++
		}
	}
	el := time.Since(t0).Seconds()
	fmt.Printf("[judgesv Go] %d juicios canon en %.3fs -> %.0f juicios/s\n", n, el, float64(n)/el)
	// 5) velocidad BFS Go
	t0 = time.Now()
	nodes := 0
	for _, v := range vs {
		_, nd := BFSSolve(MathRing{}, MState{Cur: v.a, Target: v.b, MaxSteps: 64}, 2000)
		nodes += nd
	}
	el = time.Since(t0).Seconds()
	fmt.Printf("[judgesv Go] BFS %d nodos en %.3fs -> %.0f nodos/s\n", nodes, el, float64(nodes)/el)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

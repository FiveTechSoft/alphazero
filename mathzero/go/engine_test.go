// Tests automaticos del motor (hito Q1: calidad profesional).
//
//	go test ./...        (suite completa)
//	go test -short ./... (rapida, para desarrollo)
package main

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// El juez exacto: identidades conocidas deben dar la misma forma canonica.
func TestCanonIdentidades(t *testing.T) {
	a, b, cc := V("a"), V("b"), V("c")
	iguales := []struct{ x, y *Expr }{
		// (a+b)^2 = a^2 + 2ab + b^2
		{EMul(EAdd(a, b), EAdd(a, b)),
			EAdd(EAdd(EMul(a, a), EMul(C(2), EMul(a, b))), EMul(b, b))},
		// a(b+c) = ab + ac
		{EMul(a, EAdd(b, cc)), EAdd(EMul(a, b), EMul(a, cc))},
		// a + (-a) = 0
		{EAdd(a, ENeg(a)), C(0)},
		// (a-b)(a+b) = a^2 - b^2
		{EMul(EAdd(a, ENeg(b)), EAdd(a, b)),
			EAdd(EMul(a, a), ENeg(EMul(b, b)))},
		// plegado de constantes: 2+3 = 5
		{EAdd(C(2), C(3)), C(5)},
		// 2*3*(-1) = -6
		{EMul(C(2), EMul(C(3), C(-1))), C(-6)},
		// -(a+b) = -a - b
		{ENeg(EAdd(a, b)), EAdd(ENeg(a), ENeg(b))},
	}
	for i, par := range iguales {
		if !canonEq(Canon(par.x), Canon(par.y)) {
			t.Errorf("identidad %d: el juez dice NO iguales", i)
		}
	}
	distintos := []struct{ x, y *Expr }{
		{V("a"), V("b")},
		{EAdd(a, b), EMul(a, b)},
		{C(0), C(1)},
		{EMul(a, a), a},
	}
	for i, par := range distintos {
		if canonEq(Canon(par.x), Canon(par.y)) {
			t.Errorf("no-identidad %d: el juez dice iguales", i)
		}
	}
}

// La clave estructural es invariante por conmutatividad y asociatividad.
func TestKeyNormalizacion(t *testing.T) {
	a, b, cc := V("a"), V("b"), V("c")
	ka1, ka2 := Key(EAdd(a, b)), Key(EAdd(b, a))
	if ka1 != ka2 {
		t.Errorf("conmutatividad add: %s != %s", ka1, ka2)
	}
	km1, km2 := Key(EMul(a, b)), Key(EMul(b, a))
	if km1 != km2 {
		t.Errorf("conmutatividad mul: %s != %s", km1, km2)
	}
	k1 := Key(EAdd(EAdd(a, b), cc))
	k2 := Key(EAdd(a, EAdd(b, cc)))
	k3 := Key(EAdd(cc, EAdd(a, b)))
	if k1 != k2 || k1 != k3 {
		t.Errorf("asociatividad add: %s %s %s", k1, k2, k3)
	}
	if Key(ENeg(a)) == Key(a) {
		t.Errorf("neg no debe colapsar con su argumento")
	}
}

// Toda aplicacion de toda regla preserva la forma canonica (juez exacto).
func TestReglasPreservanCanon(t *testing.T) {
	LemmaLib = nil
	exprs := []*Expr{
		EAdd(EMul(V("a"), EAdd(V("b"), C(2))), ENeg(EMul(C(3), V("c")))),
		EMul(EAdd(V("a"), ENeg(V("b"))), EAdd(V("c"), C(1))),
		ENeg(EAdd(EMul(V("a"), V("a")), EMul(C(0), V("b")))),
		EMul(EMul(C(2), V("a")), EAdd(EAdd(V("b"), C(-1)), C(3))),
		EAdd(EAdd(V("a"), V("b")), ENeg(EAdd(V("b"), V("c")))),
	}
	n := 0
	for _, e := range exprs {
		for _, act := range MathActions(e) {
			n++
			nw := applyAction(e, act)
			if !canonEq(Canon(e), Canon(nw)) {
				t.Fatalf("regla %s rompe la forma canonica", rules[act.Rule].name)
			}
		}
	}
	if n < 50 {
		t.Errorf("pocas acciones comprobadas: %d", n)
	}
}

// Get y Replace son inversos sobre todas las subposiciones.
func TestGetReplace(t *testing.T) {
	e := EAdd(EMul(V("a"), ENeg(V("b"))), EAdd(C(1), EMul(C(2), V("c"))))
	for _, p := range SubPositions(e) {
		sub := Get(e, p)
		r := Replace(e, p, sub)
		if Key(r) != Key(e) {
			t.Errorf("Get/Replace en %v no es identidad", p)
		}
		r2 := Replace(e, p, C(0))
		if canonEq(Canon(r2), Canon(e)) && size(e) > 1 {
			t.Errorf("Replace por 0 en %v no cambio nada", p)
		}
	}
}

// Regresion del juez: vectores fijados (compartidos con el port a Rust).
func TestVectoresJuez(t *testing.T) {
	f, err := os.Open("juez_vectores.txt")
	if err != nil {
		t.Skip("juez_vectores.txt no presente")
	}
	defer f.Close()
	LemmaLib = nil
	lim := 1 << 30
	if testing.Short() {
		lim = 25
	}
	sc := bufio.NewScanner(f)
	n, linea := 0, 0
	for sc.Scan() && n < lim {
		linea++
		ln := sc.Text()
		if strings.TrimSpace(ln) == "" {
			continue
		}
		parts := strings.Split(ln, " ; ")
		if len(parts) != 5 {
			t.Fatalf("linea %d mal formada", linea)
		}
		ea := parsePrefix(parts[0])
		eb := parsePrefix(parts[1])
		eqWant := parts[2] == "1"
		solWant := parts[3] == "1"
		nodWant, _ := strconv.Atoi(parts[4])
		if got := canonEq(Canon(ea), Canon(eb)); got != eqWant {
			t.Errorf("linea %d: canonEq=%v, esperado %v", linea, got, eqWant)
		}
		sol, nod := BFSSolve(MathRing{}, MState{Cur: ea, Target: eb, MaxSteps: 64}, 2000)
		if sol != solWant || nod != nodWant {
			t.Errorf("linea %d: BFS=(%v,%d), esperado (%v,%d)", linea, sol, nod, solWant, nodWant)
		}
		n++
	}
	if n == 0 {
		t.Fatal("ningun vector comprobado")
	}
	t.Logf("vectores del juez comprobados: %d", n)
}

// TicTacToe: reglas del juego sobre la interfaz generica.
func TestTTTReglas(t *testing.T) {
	g := TicTacToe{}
	s := any(TTTState{Player: 1})
	if len(g.Legal(s)) != 9 {
		t.Fatal("tablero vacio debe tener 9 legales")
	}
	if d, _ := g.Terminal(s); d {
		t.Fatal("tablero vacio no es terminal")
	}
	// X gana por la fila superior: X0 O3 X1 O4 X2
	for _, mv := range []int{0, 3, 1, 4} {
		s = g.Play(s, mv)
		if d, _ := g.Terminal(s); d {
			t.Fatal("terminal prematuro")
		}
	}
	s = g.Play(s, 2)
	done, v := g.Terminal(s)
	if !done || v != -1.0 {
		t.Errorf("victoria de X: done=%v v=%v", done, v)
	}
	// tablas
	empate := TTTState{Player: 1, Board: [9]int8{1, -1, 1, 1, -1, -1, -1, 1, 1}}
	if TTTWinner(empate) != 0 {
		t.Error("tablero de tablas no detectado")
	}
	if w := TTTWinner(TTTState{Board: [9]int8{1, 1, 1, 0, -1, 0, -1, 0, 0}}); w != 1 {
		t.Error("victoria X no detectada")
	}
}

// MCTS puro (sin red) siempre legal y determinista en tablero ganador.
func TestMCTSGanaForzada(t *testing.T) {
	g := TicTacToe{}
	// X a mover con dos en raya: debe ganar en 2 (0,1,_)
	s := any(TTTState{Player: 1, Board: [9]int8{1, 1, 0, -1, -1, 0, 0, 0, 0}})
	a := MCTS(s, g, nil, 200, 1.5)
	if a.(int) != 2 {
		t.Errorf("MCTS no elige la jugada ganadora: %v", a)
	}
}

// En un estado terminal no se buscan ni se devuelven jugadas.
func TestMCTSTerminalSinJugadas(t *testing.T) {
	g := TicTacToe{}
	done := any(TTTState{Player: 1, Board: [9]int8{1, 1, 1, -1, -1, 0, 0, 0, 0}})
	acts, visits := MCTSVisits(done, g, nil, 10, 1.5)
	if len(acts) != 0 || len(visits) != 0 {
		t.Errorf("estado terminal devolvio %d acciones y %d visitas", len(acts), len(visits))
	}
	if mv := MCTS(done, g, nil, 10, 1.5); mv != nil {
		t.Errorf("MCTS terminal devolvio %v, esperado nil", mv)
	}
}

// Una red vacia (sin acciones) no debe provocar panico en la busqueda.
func TestMCTSSinAcciones(t *testing.T) {
	acts, visits := MCTSVisits(sinAcciones{}, sinAcciones{}, nil, 5, 1.5)
	if len(acts) != 0 || len(visits) != 0 {
		t.Errorf("juego sin acciones devolvio %d acciones y %d visitas", len(acts), len(visits))
	}
}

type sinAcciones struct{}

func (sinAcciones) Legal(any) []any              { return nil }
func (sinAcciones) Play(s any, _ any) any        { return s }
func (sinAcciones) Terminal(any) (bool, float64) { return false, 0 }

// Guardar/cargar la red conserva las puntuaciones.
func TestMLPRoundTrip(t *testing.T) {
	n := NewMLP(7)
	x := make([]float64, NFeat)
	for i := range x {
		x[i] = float64(i%5) * 0.1
	}
	s1 := n.ScoreClean(x)
	tmp := t.TempDir() + "/m.json"
	n.Save(tmp)
	n2 := LoadMLP(tmp)
	if n2 == nil {
		t.Fatal("LoadMLP devolvio nil")
	}
	if s2 := n2.ScoreClean(x); math.Abs(s1-s2) > 1e-12 {
		t.Errorf("roundtrip MLP: %v vs %v", s1, s2)
	}
}

// El conjunto ciego no se solapa con el conjunto fijo de niveles.
func TestCiegoSinSolape(t *testing.T) {
	if _, err := os.Stat("blind.json"); err != nil {
		t.Skip("blind.json no presente")
	}
	keys := LoadBlindKeys("blind.json")
	if len(keys) != 700 {
		t.Errorf("conjunto ciego: %d problemas, esperados 700", len(keys))
	}
	for k := 2; k <= 8; k++ {
		for _, p := range LoadProblems("levels.json", strconv.Itoa(k)) {
			if keys[pairKey(p.Cur, p.Target)] {
				t.Errorf("solape ciego/levels en k%d", k)
			}
		}
	}
}

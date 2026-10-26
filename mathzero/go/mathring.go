// MathRing: demostracion de identidades en anillos conmutativos con
// JUEZ EXACTO (forma canonica de polinomio, coeficientes enteros).
// Acciones: (regla, posicion) aplicadas al arbol de la expresion.
// Objetivo: alcanzar la estructura del target (normalizada por
// conmutatividad/asociatividad) - hay que encontrar el camino.
package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Expr struct {
	Op   string // "const","var","add","mul","neg"
	N    int
	S    string
	A, B *Expr
}

func C(n int) *Expr         { return &Expr{Op: "const", N: n} }
func V(s string) *Expr      { return &Expr{Op: "var", S: s} }
func EAdd(a, b *Expr) *Expr { return &Expr{Op: "add", A: a, B: b} }
func EMul(a, b *Expr) *Expr { return &Expr{Op: "mul", A: a, B: b} }
func ENeg(a *Expr) *Expr    { return &Expr{Op: "neg", A: a} }

// ---- juez exacto: forma canonica monomio->coef ----

func mergeMon(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	vs := append(strings.Split(a, "*"), strings.Split(b, "*")...)
	sort.Strings(vs)
	return strings.Join(vs, "*")
}

func Canon(e *Expr) map[string]int {
	switch e.Op {
	case "const":
		if e.N == 0 {
			return map[string]int{}
		}
		return map[string]int{"": e.N}
	case "var":
		return map[string]int{e.S: 1}
	case "neg":
		r := Canon(e.A)
		for m, c := range r {
			r[m] = -c
		}
		return r
	case "add":
		r := Canon(e.A)
		for m, c := range Canon(e.B) {
			r[m] += c
			if r[m] == 0 {
				delete(r, m)
			}
		}
		return r
	case "mul":
		r := map[string]int{}
		for m1, c1 := range Canon(e.A) {
			for m2, c2 := range Canon(e.B) {
				m := mergeMon(m1, m2)
				r[m] += c1 * c2
				if r[m] == 0 {
					delete(r, m)
				}
			}
		}
		return r
	}
	return nil
}

func canonEq(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for m, c := range a {
		if b[m] != c {
			return false
		}
	}
	return true
}

// ---- clave estructural normalizada ----

var keyCache = map[*Expr]string{}

func ClearKeyCache() { keyCache = map[*Expr]string{} }

func Key(e *Expr) string {
	if k, ok := keyCache[e]; ok {
		return k
	}
	k := keyCompute(e)
	keyCache[e] = k
	return k
}

func keyCompute(e *Expr) string {
	switch e.Op {
	case "const":
		return fmt.Sprintf("c%d", e.N)
	case "var":
		return "v" + e.S
	case "neg":
		return "neg(" + Key(e.A) + ")"
	}
	parts := []string{}
	var rec func(x *Expr)
	rec = func(x *Expr) {
		if x.Op == e.Op {
			rec(x.A)
			rec(x.B)
		} else {
			parts = append(parts, Key(x))
		}
	}
	rec(e)
	sort.Strings(parts)
	return e.Op + "(" + strings.Join(parts, ",") + ")"
}

// ---- posiciones y sustitucion ----

func SubPositions(e *Expr) [][]int {
	out := [][]int{{}}
	var rec func(x *Expr, p []int)
	rec = func(x *Expr, p []int) {
		if x.Op == "add" || x.Op == "mul" {
			out = append(out, append(append([]int{}, p...), 0))
			out = append(out, append(append([]int{}, p...), 1))
			rec(x.A, append(append([]int{}, p...), 0))
			rec(x.B, append(append([]int{}, p...), 1))
		} else if x.Op == "neg" {
			out = append(out, append(append([]int{}, p...), 0))
			rec(x.A, append(append([]int{}, p...), 0))
		}
	}
	rec(e, []int{})
	return out
}

func Get(e *Expr, p []int) *Expr {
	for _, i := range p {
		if i == 0 {
			e = e.A
		} else {
			e = e.B
		}
	}
	return e
}

func Replace(e *Expr, p []int, nw *Expr) *Expr {
	if len(p) == 0 {
		return nw
	}
	if e.Op == "neg" {
		return ENeg(Replace(e.A, p[1:], nw))
	}
	if p[0] == 0 {
		return &Expr{Op: e.Op, A: Replace(e.A, p[1:], nw), B: e.B}
	}
	return &Expr{Op: e.Op, A: e.A, B: Replace(e.B, p[1:], nw)}
}

// ---- reglas (mismas 14 que Python) ----

func is0(x *Expr) bool { return x.Op == "const" && x.N == 0 }
func is1(x *Expr) bool { return x.Op == "const" && x.N == 1 }

var rules = []struct {
	name string
	f, r func(*Expr) *Expr
}{
	{"add0",
		func(x *Expr) *Expr {
			if x.Op == "add" && is0(x.B) {
				return x.A
			}
			if x.Op == "add" && is0(x.A) {
				return x.B
			}
			return nil
		},
		func(x *Expr) *Expr { return EAdd(x, C(0)) }},
	{"mul1",
		func(x *Expr) *Expr {
			if x.Op == "mul" && is1(x.B) {
				return x.A
			}
			if x.Op == "mul" && is1(x.A) {
				return x.B
			}
			return nil
		},
		func(x *Expr) *Expr { return EMul(x, C(1)) }},
	{"mul0",
		func(x *Expr) *Expr {
			if x.Op == "mul" && (is0(x.A) || is0(x.B)) {
				return C(0)
			}
			return nil
		}, nil},
	{"comm+",
		func(x *Expr) *Expr {
			if x.Op == "add" {
				return EAdd(x.B, x.A)
			}
			return nil
		}, nil},
	{"comm*",
		func(x *Expr) *Expr {
			if x.Op == "mul" {
				return EMul(x.B, x.A)
			}
			return nil
		}, nil},
	{"assoc+",
		func(x *Expr) *Expr {
			if x.Op == "add" && x.A.Op == "add" {
				return EAdd(x.A.A, EAdd(x.A.B, x.B))
			}
			return nil
		},
		func(x *Expr) *Expr {
			if x.Op == "add" && x.B.Op == "add" {
				return EAdd(EAdd(x.A, x.B.A), x.B.B)
			}
			return nil
		}},
	{"assoc*",
		func(x *Expr) *Expr {
			if x.Op == "mul" && x.A.Op == "mul" {
				return EMul(x.A.A, EMul(x.A.B, x.B))
			}
			return nil
		},
		func(x *Expr) *Expr {
			if x.Op == "mul" && x.B.Op == "mul" {
				return EMul(EMul(x.A, x.B.A), x.B.B)
			}
			return nil
		}},
	{"dist",
		func(x *Expr) *Expr {
			if x.Op == "mul" && x.B.Op == "add" {
				return EAdd(EMul(x.A, x.B.A), EMul(x.A, x.B.B))
			}
			return nil
		},
		func(x *Expr) *Expr {
			if x.Op == "add" && x.A.Op == "mul" && x.B.Op == "mul" && Key(x.A.A) == Key(x.B.A) {
				return EMul(x.A.A, EAdd(x.A.B, x.B.B))
			}
			return nil
		}},
	{"negneg",
		func(x *Expr) *Expr {
			if x.Op == "neg" && x.A.Op == "neg" {
				return x.A.A
			}
			return nil
		},
		func(x *Expr) *Expr { return ENeg(ENeg(x)) }},
	{"neg+",
		func(x *Expr) *Expr {
			if x.Op == "neg" && x.A.Op == "add" {
				return EAdd(ENeg(x.A.A), ENeg(x.A.B))
			}
			return nil
		},
		func(x *Expr) *Expr {
			if x.Op == "add" && x.A.Op == "neg" && x.B.Op == "neg" {
				return ENeg(EAdd(x.A.A, x.B.A))
			}
			return nil
		}},
	{"+-0",
		func(x *Expr) *Expr {
			if x.Op == "add" && x.B.Op == "neg" && Key(x.A) == Key(x.B.A) {
				return C(0)
			}
			if x.Op == "add" && x.A.Op == "neg" && Key(x.B) == Key(x.A.A) {
				return C(0)
			}
			return nil
		}, nil},
	{"neg*",
		func(x *Expr) *Expr {
			if x.Op == "neg" && x.A.Op == "mul" {
				return EMul(ENeg(x.A.A), x.A.B)
			}
			return nil
		},
		func(x *Expr) *Expr {
			if x.Op == "mul" && x.A.Op == "neg" {
				return ENeg(EMul(x.A.A, x.B))
			}
			return nil
		}},
	{"fold+",
		func(x *Expr) *Expr {
			if x.Op == "add" && x.A.Op == "const" && x.B.Op == "const" {
				return C(x.A.N + x.B.N)
			}
			return nil
		}, nil},
	{"fold*",
		func(x *Expr) *Expr {
			if x.Op == "mul" && x.A.Op == "const" && x.B.Op == "const" {
				return C(x.A.N * x.B.N)
			}
			return nil
		}, nil},
}

type MAction struct {
	Rule  int // indice en rules; -1 = accion de lema
	Pos   []int
	Lemma int  // indice en LemmaLib.Lemmas; -1 = regla normal
	Dir   int8 // E11: 0 = f, 1 = r; -1 = sin direccion (lema o motor clasico)
}

// E11: si true, MathActions fija Dir y deduplica sucesores identicos, y
// applyAction usa Dir sin reprobar (bug medido: 30% acciones duplicadas,
// 1.044 direcciones r enmascaradas por la reprobacion f-primero).
var DirFix bool

func MathActions(e *Expr) []MAction {
	var out []MAction
	seenSucc := map[string]bool{}
	for _, p := range SubPositions(e) {
		sub := Get(e, p)
		for i, rl := range rules {
			for d, fn := range []func(*Expr) *Expr{rl.f, rl.r} {
				if fn == nil {
					continue
				}
				nw := fn(sub)
				if nw != nil && Key(nw) != Key(sub) {
					if DirFix {
						kf := Key(Replace(e, p, nw))
						if seenSucc[kf] {
							continue
						}
						seenSucc[kf] = true
						out = append(out, MAction{Rule: i, Pos: p, Lemma: -1, Dir: int8(d)})
					} else {
						out = append(out, MAction{Rule: i, Pos: p, Lemma: -1, Dir: -1})
					}
				}
			}
		}
	}
	// acciones de lema: reescrituras ya demostradas, en un solo paso
	if LemmaLib != nil {
		for _, p := range SubPositions(e) {
			ks := Key(Get(e, p))
			for _, li := range LemmaLib.ByKey[ks] {
				if DirFix {
					lm := LemmaLib.Lemmas[li]
					nw := lm.To
					if ks != lm.FromKey {
						nw = lm.From
					}
					kf := Key(Replace(e, p, nw))
					if seenSucc[kf] {
						continue
					}
					seenSucc[kf] = true
				}
				out = append(out, MAction{Rule: -1, Pos: p, Lemma: li, Dir: -1})
			}
		}
	}
	return out
}

func applyAction(e *Expr, a MAction) *Expr {
	if a.Lemma >= 0 {
		lm := LemmaLib.Lemmas[a.Lemma]
		if Key(Get(e, a.Pos)) == lm.FromKey {
			return Replace(e, a.Pos, lm.To)
		}
		return Replace(e, a.Pos, lm.From)
	}
	sub := Get(e, a.Pos)
	rl := rules[a.Rule]
	if a.Dir >= 0 {
		fn := rl.f
		if a.Dir == 1 {
			fn = rl.r
		}
		if fn == nil {
			return e
		}
		nw := fn(sub)
		if nw != nil && Key(nw) != Key(sub) {
			return Replace(e, a.Pos, nw)
		}
		return e
	}
	for _, fn := range []func(*Expr) *Expr{rl.f, rl.r} {
		if fn == nil {
			continue
		}
		nw := fn(sub)
		if nw != nil && Key(nw) != Key(sub) {
			return Replace(e, a.Pos, nw)
		}
	}
	return e
}

// ---- MathRing como Game ----

type MState struct {
	Cur, Target *Expr
	Steps       int
	MaxSteps    int
}

type MathRing struct{}

func (MathRing) Legal(s any) []any {
	st := s.(MState)
	acts := MathActions(st.Cur)
	out := make([]any, len(acts))
	for i, a := range acts {
		out[i] = a
	}
	return out
}

func (MathRing) Play(s any, a any) any {
	st := s.(MState)
	st.Cur = applyAction(st.Cur, a.(MAction))
	st.Steps++
	return st
}

func (MathRing) Terminal(s any) (bool, float64) {
	st := s.(MState)
	if Key(st.Cur) == Key(st.Target) {
		return true, 1.0
	}
	if st.Steps >= st.MaxSteps {
		return true, -1.0
	}
	return false, 0
}

// ---- problemas (mismo conjunto fijo que Python) ----

func LoadProblems(path, k string) []MState {
	data := readDataFile(path)
	if data == nil {
		fmt.Println("aviso: no se encuentra " + path + " (ni en disco ni embebido); se omite")
		return nil
	}
	var raw map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(err)
	}
	var out []MState
	for _, pr := range raw[k] {
		out = append(out, MState{Cur: parseExpr(pr["a"]), Target: parseExpr(pr["b"]), MaxSteps: 2 * len(k) * 4})
	}
	return out
}

func parseExpr(m json.RawMessage) *Expr {
	var d map[string]json.RawMessage
	if err := json.Unmarshal(m, &d); err != nil {
		panic(err)
	}
	var op string
	json.Unmarshal(d["op"], &op)
	switch op {
	case "const":
		var n int
		json.Unmarshal(d["n"], &n)
		return C(n)
	case "var":
		var s string
		json.Unmarshal(d["s"], &s)
		return V(s)
	case "neg":
		return ENeg(parseExpr(d["a"]))
	default:
		return &Expr{Op: op, A: parseExpr(d["a"]), B: parseExpr(d["b"])}
	}
}

// ---- baseline BFS (sin red) ----

func BFSSolve(g MathRing, p MState, cap int) (bool, int) {
	ClearKeyCache()
	target := Key(p.Target)
	seen := map[string]bool{Key(p.Cur): true}
	queue := []*Expr{p.Cur}
	nodes := 0
	for len(queue) > 0 && nodes < cap {
		cur := queue[0]
		queue = queue[1:]
		nodes++
		if Key(cur) == target {
			return true, nodes
		}
		for _, a := range MathActions(cur) {
			nxt := applyAction(cur, a)
			k2 := Key(nxt)
			if !seen[k2] {
				seen[k2] = true
				queue = append(queue, nxt)
			}
		}
	}
	return false, nodes
}

// ---- chequeo del juez: toda regla preserva la forma canonica ----

func CheckRules(g MathRing, n int) {
	// bateria pequena: todas las reglas aplicadas a cada nodo de un conjunto de expresiones
	bad := 0
	tested := 0
	exprs := []*Expr{
		EAdd(EMul(V("a"), EAdd(V("b"), C(2))), ENeg(EMul(C(3), V("c")))),
		EMul(EAdd(V("a"), ENeg(V("b"))), EAdd(V("c"), C(1))),
		ENeg(EAdd(EMul(V("a"), V("a")), EMul(C(0), V("b")))),
		EMul(EMul(C(2), V("a")), EAdd(EAdd(V("b"), C(-1)), C(3))),
		EAdd(EAdd(V("a"), V("b")), ENeg(EAdd(V("b"), V("c")))),
	}
	for _, e := range exprs {
		for _, a := range MathActions(e) {
			tested++
			nw := applyAction(e, a)
			if !canonEq(Canon(e), Canon(nw)) {
				bad++
				fmt.Println("REGLA ROTA:", rules[a.Rule].name)
			}
		}
	}
	fmt.Printf("[MathRing] juez exacto: %d aplicaciones de reglas verificadas, fallos = %d\n", tested, bad)
}

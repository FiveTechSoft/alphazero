package main

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

// Diagnostico de la revision externa: MAction no guarda direccion (f/r).
// MathActions emite UNA accion por direccion aplicable, pero applyAction
// vuelve a probar f primero: cuando ambas direcciones aplican y cambian,
// las dos acciones colapsan al sucesor de f y la r pretendida se pierde.
// Este test NO modifica el motor: solo mide.
func TestDirectionDiag(t *testing.T) {
	exprs := []*Expr{
		EAdd(EAdd(V("a"), V("b")), EAdd(V("c"), V("d"))), // assoc+ ambas
		EMul(EMul(V("a"), V("b")), EMul(V("c"), V("d"))), // assoc* ambas
		EAdd(V("a"), C(0)), // add0 f y r
		EMul(V("a"), C(1)), // mul1 f y r
		EMul(EAdd(V("a"), C(0)), EAdd(V("b"), V("c"))),
		ENeg(ENeg(EAdd(V("x"), C(0)))),
	}
	r := rand.New(rand.NewSource(7))
	for i := 0; i < 300; i++ { // stream tipo genTrain k=5
		a, b := genTrain(r, 5)
		exprs = append(exprs, a, b)
	}
	totalActs, dupActs, misActs := 0, 0, 0
	bothDir := map[string]int{}
	dupByRule := map[string]int{}
	misByRule := map[string]int{}
	examples := map[string]string{}
	for _, e := range exprs {
		acts := MathActions(e)
		totalActs += len(acts)
		seenSucc := map[string]bool{}
		for _, a := range acts {
			if a.Rule < 0 {
				continue
			}
			k := Key(applyAction(e, a))
			if seenSucc[k] {
				dupActs++
				dupByRule[rules[a.Rule].name]++
			}
			seenSucc[k] = true
		}
		// recorremos posiciones/reglas con direccion explicita
		for _, p := range SubPositions(e) {
			sub := Get(e, p)
			for i, rl := range rules {
				f, rr := rl.f, rl.r
				fNw := maybeApply(f, sub)
				rNw := maybeApply(rr, sub)
				fOk := fNw != nil && Key(fNw) != Key(sub)
				rOk := rNw != nil && Key(rNw) != Key(sub)
				if fOk && rOk {
					bothDir[rl.name]++
					// la accion emitida para la direccion r (Rule=i, Pos=p):
					// applyAction la resuelve como f -> sucesor equivocado
					real := Key(applyAction(e, MAction{Rule: i, Pos: p, Lemma: -1}))
					if real != Key(rNw) {
						misActs++
						misByRule[rl.name]++
						if examples[rl.name] == "" {
							examples[rl.name] = fmt.Sprintf("sub=%s | f->%s | r(pretendido)->%s | applyAction->%s",
								Key(sub), Key(fNw), Key(rNw), real)
						}
					}
				}
			}
		}
	}
	fmt.Printf("[diag] expresiones=%d acciones_emitidas=%d duplicadas_mismo_sucesor=%d (%.2f%%)\n",
		len(exprs), totalActs, dupActs, 100*float64(dupActs)/float64(totalActs))
	fmt.Printf("[diag] acciones con sucesor equivocado (r pretendida, sale f)=%d\n", misActs)
	names := []string{}
	for n := range bothDir {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("[diag] regla %-7s ambas_direcciones=%d mal_aplicadas=%d dup=%d\n", n, bothDir[n], misByRule[n], dupByRule[n])
	}
	for _, n := range names {
		fmt.Printf("[diag] ejemplo %s: %s\n", n, examples[n])
	}
}

func maybeApply(fn func(*Expr) *Expr, e *Expr) *Expr {
	if fn == nil {
		return nil
	}
	return fn(e)
}

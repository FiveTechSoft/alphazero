// EvalFail: analisis de fallos de la red. Para cada nivel, evalua la red
// (GuidedSolve, capNet) y, con un oraculo BFS (capOr), clasifica cada fallo:
// profundidad minima de la solucion, reglas del camino oraculo, y ejemplos.
//
//	./mathzero evalfail [levels.json|blind.json] [modelo] [capNet] [capOr] [kmin] [kmax]
package main

import (
	"fmt"
	"os"
	"sort"
)

// BFSOraculo: BFS con padres; devuelve (resuelto, nodos, profundidad, reglas del camino).
func BFSOraculo(start, target *Expr, cap int) (bool, int, int, []string) {
	ClearKeyCache()
	tk := Key(target)
	type node struct {
		e   *Expr
		dep int
		par int
		act MAction
	}
	pool := []node{{e: start, dep: 0, par: -1}}
	seen := map[string]bool{Key(start): true}
	head, nodes := 0, 0
	for head < len(pool) && nodes < cap {
		cur := pool[head]
		head++
		nodes++
		if Key(cur.e) == tk {
			var ru []string
			for x := cur; x.par >= 0; x = pool[x.par] {
				ru = append(ru, RuleName(x.act))
			}
			return true, nodes, cur.dep, ru
		}
		for _, a := range MathActions(cur.e) {
			nx := applyAction(cur.e, a)
			k := Key(nx)
			if !seen[k] {
				seen[k] = true
				pool = append(pool, node{e: nx, dep: cur.dep + 1, par: head - 1, act: a})
			}
		}
	}
	return false, nodes, -1, nil
}

func histStr(h map[int]int) string {
	var ks []int
	for k := range h {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	s := ""
	for _, k := range ks {
		s += fmt.Sprintf(" %d:%d", k, h[k])
	}
	return s
}

func EvalFail(args []string) {
	var ok bool
	if args, ok = parseDataFlags(args); !ok {
		return
	}
	if len(args) > 6 {
		fmt.Fprintln(os.Stderr, "uso: mathzero evalfail [datos] [modelo] [capNet] [capOr] [kmin] [kmax] [-datadir DIR]")
		return
	}
	path, model := "levels.json", "model_math.json"
	capNet, capOr, kmin, kmax := 2000, 50000, 2, 8
	if len(args) > 0 {
		path = args[0]
	}
	if len(args) > 1 {
		model = args[1]
	}
	if len(args) > 2 {
		var ok bool
		if capNet, ok = parseIntArg("capNet", args[2], 1); !ok {
			return
		}
	}
	if len(args) > 3 {
		var ok bool
		if capOr, ok = parseIntArg("capOr", args[3], 1); !ok {
			return
		}
	}
	if len(args) > 4 {
		var ok bool
		if kmin, ok = parseIntArg("kmin", args[4], 0); !ok {
			return
		}
	}
	if len(args) > 5 {
		var ok bool
		if kmax, ok = parseIntArg("kmax", args[5], 0); !ok {
			return
		}
	}
	if kmin > kmax {
		fmt.Fprintln(os.Stderr, "argumento invalido: kmin no puede ser mayor que kmax")
		return
	}
	net := LoadMLP(model)
	if net == nil {
		fmt.Println("[evalfail] sin modelo", model)
		return
	}
	LemmaLib = LoadLibrary("lemas.json")
	fmt.Printf("[evalfail] datos=%s modelo=%s(nivel %d ronda %d) capRed=%d capOraculo=%d lemas=%d\n",
		path, model, net.Level, net.Round, capNet, capOr, len(LemmaLib.Lemmas))
	totOK, tot := 0, 0
	for k := kmin; k <= kmax; k++ {
		probs := LoadProblems(path, fmt.Sprint(k))
		ok, orFail := 0, 0
		depOK, depFail := map[int]int{}, map[int]int{}
		ruleCount := map[string]int{}
		var ejemplos []string
		for _, p := range probs {
			s, _, _ := GuidedSolve(net, p.Cur, p.Target, capNet)
			so, _, dep, ru := BFSOraculo(p.Cur, p.Target, capOr)
			if s {
				ok++
				if so {
					depOK[dep]++
				}
			} else if !so {
				orFail++
			} else {
				depFail[dep]++
				for _, rn := range ru {
					ruleCount[rn]++
				}
				if len(ejemplos) < 3 {
					ejemplos = append(ejemplos, fmt.Sprintf("    dep=%d %s  =>  %s", dep, RenderPrefix(p.Cur), RenderPrefix(p.Target)))
				}
			}
		}
		totOK += ok
		tot += len(probs)
		fmt.Printf("[k%d] red %d/%d | profundidad(resueltos):%s | fallos: profundidad%s | oraculo_tampoco=%d\n",
			k, ok, len(probs), histStr(depOK), histStr(depFail), orFail)
		var rs []string
		for rn, c := range ruleCount {
			rs = append(rs, fmt.Sprintf("%s:%d", rn, c))
		}
		sort.Strings(rs)
		if len(rs) > 0 {
			fmt.Printf("  reglas en caminos oraculo de fallos: %v\n", rs)
		}
		for _, e := range ejemplos {
			fmt.Println(e)
		}
	}
	fmt.Printf("[evalfail] TOTAL red %d/%d\n", totOK, tot)
}

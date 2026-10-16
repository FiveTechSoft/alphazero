// Subcomando CLI de geometria (opt-in). NO forma parte del curriculo por
// defecto, no entrena, no toca el modelo ni la biblioteca. Carga el banco de
// identidades geometricas, verifica cada una con el JUEZ EXACTO (canonEq) e
// intenta ademas la busqueda guiada con un presupuesto de nodos.
//
//	./mathzero geometry [presupuesto_nodos]
package main

import (
	"fmt"
	"strconv"
)

func yesno(b bool) string {
	if b {
		return "si"
	}
	return "no"
}

func GeomMain(args []string) {
	budget := 2000
	if len(args) > 0 {
		if b, err := strconv.Atoi(args[0]); err == nil && b > 0 {
			budget = b
		}
	}
	const levels = 4
	fmt.Printf("[geometry] banco opt-in de identidades; presupuesto de busqueda=%d nodos\n", budget)
	valid, solved, total := 0, 0, 0
	for lvl := 1; lvl <= levels; lvl++ {
		probs := LoadGeometryProblems(strconv.Itoa(lvl), budget)
		if len(probs) == 0 {
			fmt.Printf("[geometry] nivel %d: sin problemas\n", lvl)
			continue
		}
		for i, p := range probs {
			total++
			ident := canonEq(Canon(p.Cur), Canon(p.Target))
			ok, nodes, sm := GuidedSolve(nil, p.Cur, p.Target, budget)
			if ident {
				valid++
			}
			if ok {
				solved++
			}
			fmt.Printf("[geometry] nivel %d prob %d: juez_exacto=%s busqueda=%s nodos=%d/%d pasos=%d size=%d\n",
				lvl, i, yesno(ident), yesno(ok), nodes, budget, len(sm), size(p.Cur))
		}
	}
	fmt.Printf("[geometry] juez exacto (canonEq): %d/%d identidades validas\n", valid, total)
	fmt.Printf("[geometry] busqueda guiada (uniforme, cap=%d): %d/%d probadas\n", budget, solved, total)
}

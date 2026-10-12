// Render de expresiones a texto legible y nombres de acciones.
package main

import (
	"encoding/json"
	"fmt"
)

func Render(e *Expr) string {
	switch e.Op {
	case "const":
		return fmt.Sprint(e.N)
	case "var":
		return e.S
	case "neg":
		return "-" + renderAtom(e.A)
	case "mul":
		return renderAtom(e.A) + "*" + renderAtom(e.B)
	case "add":
		return Render(e.A) + " + " + Render(e.B)
	}
	return "?"
}

// renderAtom pone parentesis a todo lo que no sea hoja o negativo.
func renderAtom(e *Expr) string {
	if e.Op == "const" || e.Op == "var" || e.Op == "neg" || e.Op == "mul" {
		return Render(e)
	}
	return "(" + Render(e) + ")"
}

// RuleName nombra una accion (regla o lema).
func RuleName(a MAction) string {
	if a.Lemma >= 0 {
		return fmt.Sprintf("lema#%d", a.Lemma)
	}
	if a.Rule >= 0 && a.Rule < len(rules) {
		return rules[a.Rule].name
	}
	return "?"
}

// LoadProblemsBytes: mismo formato que LoadProblems pero desde bytes
// (para el build WebAssembly, sin sistema de ficheros).
func LoadProblemsBytes(data []byte, k string) []MState {
	var raw map[string][]map[string]json.RawMessage
	if json.Unmarshal(data, &raw) != nil {
		return nil
	}
	var out []MState
	for _, pr := range raw[k] {
		out = append(out, MState{Cur: parseExpr(pr["a"]), Target: parseExpr(pr["b"]), MaxSteps: 2 * len(k) * 4})
	}
	return out
}

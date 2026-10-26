// Descubrimiento (h15): genera problemas fuera del curriculum y del ciego,
// los resuelve con el motor y se queda con identidades NUEVAS que no son
// una sola aplicacion de las 14 reglas. El juez exacto (canonEq) valida cada
// lema antes de insertarlo en la biblioteca.
//
//	mathzero discover [n] [cap] [semilla] [-json] [-nosave] [-datadir DIR]
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

type discovered struct {
	From   map[string]any `json:"from"`
	To     map[string]any `json:"to"`
	Pasos  int            `json:"pasos"`
	Saving int            `json:"saving"`
}

func isSingleRule(a, b *Expr) bool {
	kb := Key(b)
	for _, act := range MathActions(a) {
		nx := applyAction(a, act)
		if Key(nx) == kb {
			return true
		}
	}
	return false
}

func DiscoverMain(args []string) {
	var ok bool
	if args, ok = parseDataFlags(args); !ok {
		return
	}
	wantJSON := false
	var pos []string
	for _, a := range args {
		if a == "-json" {
			wantJSON = true
		} else {
			pos = append(pos, a)
		}
	}
	if len(pos) > 3 {
		fmt.Fprintln(os.Stderr, "uso: mathzero discover [n] [cap] [semilla] [-json] [-nosave] [-datadir DIR]")
		return
	}
	n, cap, seed := 5, 2000, int64(1)
	if len(pos) > 0 {
		if n, ok = parseIntArg("n", pos[0], 1); !ok {
			return
		}
	}
	if len(pos) > 1 {
		if cap, ok = parseIntArg("cap", pos[1], 1); !ok {
			return
		}
	}
	if len(pos) > 2 {
		if seed, ok = parseInt64Arg("semilla", pos[2]); !ok {
			return
		}
	}
	DirFix = true
	net := LoadMLP(dataPath("model_math.json"))
	if net == nil {
		fmt.Fprintln(os.Stderr, "[discover] no se pudo cargar el modelo")
		return
	}
	lib := LoadLibrary(dataPath("lemas.json"))
	LemmaLib = nil
	used := map[string]bool{}
	for _, f := range []string{"levels.json", "blind.json"} {
		for k := 2; k <= 8; k++ {
			for _, p := range LoadProblems(dataPath(f), fmt.Sprint(k)) {
				used[pairKey(p.Cur, p.Target)] = true
			}
		}
	}
	r := rand.New(rand.NewSource(seed))
	var out []discovered
	for i := 0; i < n*20 && len(out) < n; i++ {
		a, b := genTrain(r, 8)
		if used[pairKey(a, b)] {
			continue
		}
		okSolve, _, sm := GuidedSolve(net, a, b, cap)
		if !okSolve || len(sm) < 2 || isSingleRule(a, b) || lib.Contains(a, b) {
			continue
		}
		if !canonEq(Canon(a), Canon(b)) {
			continue
		}
		if lib.Insert(a, b, len(sm)) {
			out = append(out, discovered{From: exprJSON(a), To: exprJSON(b),
				Pasos: len(sm), Saving: len(sm)})
		}
	}
	for _, d := range out {
		fmt.Printf("[discover] nuevo lema (%d pasos): %v -> %v\n", d.Pasos, d.From, d.To)
	}
	fmt.Printf("[discover] %d lemas nuevos verificados (semilla %d, cap %d, biblioteca %d)\n",
		len(out), seed, cap, len(lib.Lemmas))
	if len(out) > 0 {
		LemmaLib = lib
		saveLibrary(lib, dataPath("lemas.json"))
	}
	if wantJSON {
		data, _ := json.MarshalIndent(out, "", "  ")
		name := dataPath(fmt.Sprintf("descubrimientos_seed%d.json", seed))
		if err := saveJSON(name, data); err != nil {
			fmt.Fprintln(os.Stderr, "[discover] error escribiendo", name, ":", err)
			return
		}
		if !NoSave {
			fmt.Println("[discover] JSON escrito en", name)
		}
	}
}

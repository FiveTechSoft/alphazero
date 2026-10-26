// EvalBlind: evaluacion SOLO-lectura del pipeline shipped sobre el
// conjunto ciego. No entrena y no escribe model_math.json/lemas.json.
//
//	mathzero evalblind [cap] [modelo]
package main

import (
	"fmt"
	"os"
)

func EvalBlindMain(args []string) {
	if len(args) > 2 {
		fmt.Fprintln(os.Stderr, "uso: mathzero evalblind [cap] [modelo]")
		return
	}
	cap, model := 1000, "model_math.json"
	if len(args) > 0 {
		var ok bool
		if cap, ok = parseIntArg("cap", args[0], 1); !ok {
			return
		}
	}
	if len(args) > 1 {
		model = args[1]
	}
	net := LoadMLP(model)
	if net == nil {
		fmt.Fprintln(os.Stderr, "[evalblind] no se pudo cargar el modelo:", model)
		return
	}
	DirFix = true
	LemmaLib = nil
	lv := map[int][]MState{}
	for k := 2; k <= 8; k++ {
		lv[k] = LoadProblems("blind.json", fmt.Sprint(k))
	}
	e := evalBlind(net, lv, cap)
	pl := 0.0
	if e.nSolved > 0 {
		pl = float64(e.proofs) / float64(e.nSolved)
	}
	fmt.Printf("[evalblind] modelo=%s nivel=%d ronda=%d cap=%d lemas=0\n",
		model, net.Level, net.Round, cap)
	fmt.Printf("[evalblind] k2:%d k3:%d k4:%d k5:%d k6:%d k7:%d k8:%d total=%d/700 nodos_med=%d prueba_med=%.1f\n",
		e.perLevel[2], e.perLevel[3], e.perLevel[4], e.perLevel[5],
		e.perLevel[6], e.perLevel[7], e.perLevel[8], e.total, e.nodes, pl)
}

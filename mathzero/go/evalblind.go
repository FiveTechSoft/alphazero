// EvalBlind: evaluacion SOLO-lectura del pipeline shipped sobre el
// conjunto ciego. No entrena y no escribe model_math.json/lemas.json.
//
//	mathzero evalblind [cap] [modelo]
package main

import (
	"fmt"
	"os"
)

// evalBlindFreq: como evalBlind pero con feedback de frecuencia
// (freqrule de he_pilot3): cada problema resuelto alimenta el corpus
// y los hijos ya vistos en pruebas cuestan menos (lam * log1p(freq)).
func evalBlindFreq(net *MLP, lv map[int][]MState, cap int, lam float64) blindEval {
	var r blindEval
	ClearKeyCache()
	FreqT = map[string]int{}
	FreqLam = lam
	defer func() { FreqT, FreqLam, LastPathKeys = nil, 0, nil; ClearKeyCache() }()
	for k := 2; k <= 8; k++ {
		nd := 0
		for _, p := range lv[k] {
			ok, n, sm := GuidedSolve(net, p.Cur, p.Target, cap)
			nd += n
			if ok {
				r.perLevel[k]++
				r.total++
				r.proofs += len(sm)
				r.nSolved++
				AddPathFreqs(LastPathKeys)
			}
		}
		r.nodes += nd / len(lv[k])
	}
	r.nodes /= 7
	return r
}

func EvalBlindMain(args []string) {
	var ok bool
	if args, ok = parseDataFlags(args); !ok {
		return
	}
	if len(args) > 3 {
		fmt.Fprintln(os.Stderr, "uso: mathzero evalblind [cap] [modelo] [lam] [-datadir DIR]")
		return
	}
	cap, model := 1000, dataPath("model_math.json")
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
		lv[k] = LoadProblems(dataPath("blind.json"), fmt.Sprint(k))
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
	if len(args) > 2 {
		lam, ok2 := parseFloatArg("lam", args[2])
		if !ok2 {
			return
		}
		ef := evalBlindFreq(net, lv, cap, lam)
		plf := 0.0
		if ef.nSolved > 0 {
			plf = float64(ef.proofs) / float64(ef.nSolved)
		}
		fmt.Printf("[evalblind+freq] lam=%g corpus online (feedback de pruebas resueltas)\n", lam)
		fmt.Printf("[evalblind+freq] k2:%d k3:%d k4:%d k5:%d k6:%d k7:%d k8:%d total=%d/700 nodos_med=%d prueba_med=%.1f\n",
			ef.perLevel[2], ef.perLevel[3], ef.perLevel[4], ef.perLevel[5],
			ef.perLevel[6], ef.perLevel[7], ef.perLevel[8], ef.total, ef.nodes, plf)
		fmt.Printf("[evalblind+freq] DELTA vs baseline: %d/700\n", ef.total-e.total)
	}
}

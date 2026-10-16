// Ciclo auto-mejorante medido (hito 6/7 del ROADMAP):
//
//	autojuego -> entrenamiento -> verificacion con juez exacto ->
//	biblioteca persistente de lemas -> metrica repetible por ciclo.
//
// La metrica es FIJA (levels.json) y el comando es unico:
//
//	./mathzero cycle [ciclos] [problemas_por_ciclo] [cap]
//
// Imprime una tabla por ciclo: resueltos por nivel (conjunto fijo de
// evaluacion) y nodos medios, con y sin biblioteca (ablacion en el
// primer y ultimo ciclo), asi la curva de crecimiento es repetible.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
)

func evalAll(net *MLP, lv map[int][]MState, cap int) (string, int) {
	savedTraining := SearchTraining
	SearchTraining = false
	defer func() { SearchTraining = savedTraining }()
	s := ""
	tot := 0
	for k := 2; k <= 8; k++ {
		o, nd := EvalLevel(net, lv[k], cap)
		s += fmt.Sprintf("k%d:%d/%d(%d) ", k, o, len(lv[k]), nd)
		tot += o
	}
	return s, tot
}

func Cycle(nCycles, perRound, cap int, modelPath, libPath string, seed int64) {
	if nCycles < 1 || perRound < 1 || cap < 2 {
		fmt.Println("[cycle] invalid cycle/perRound/cap")
		return
	}
	r := rand.New(rand.NewSource(seed))
	net := LoadMLP(modelPath)
	if net == nil {
		net = NewMLP(seed)
		fmt.Println("[cycle] red nueva")
	} else {
		fmt.Printf("[cycle] continuando desde %s (nivel %d, ronda %d)\n", modelPath, net.Level, net.Round)
	}
	LemmaLib = LoadLibrary(libPath)
	fmt.Printf("[cycle] biblioteca: %d lemas cargados\n", len(LemmaLib.Lemmas))
	lv := map[int][]MState{}
	for k := 2; k <= 8; k++ {
		lv[k] = LoadLevels(k)
	}
	// Fixed evaluation is never used for labels, utility, hindsight or generation.
	fixedPairs := map[string]bool{}
	for k := 2; k <= 8; k++ {
		if len(lv[k]) == 0 {
			fmt.Println("[cycle] empty fixed evaluation level", k)
			return
		}
		for _, p := range lv[k] {
			fixedPairs[RenderPrefix(p.Cur)+"|"+RenderPrefix(p.Target)] = true
			fixedPairs[RenderPrefix(p.Target)+"|"+RenderPrefix(p.Cur)] = true
		}
	}
	fmt.Printf("[cycle] utility lambda=%g min_problems=%d active_cap=%d; bidirectional/hindsight enabled\n", LemmaLib.UtilityLambda, LemmaLib.UtilityMinProblems, LemmaLib.ActiveCap)
	fmt.Println("[cycle] metrica fija: levels.json, cap", cap, "(resueltos y nodos medios por nivel)")
	bestTot := -1
	for c := 1; c <= nCycles; c++ {
		// 1) METRICA del estado actual (antes de aprender mas)
		metric, tot := evalAll(net, lv, cap)
		if tot > bestTot {
			bestTot = tot
			net.Save(modelPath + ".best")
			fmt.Printf("[ciclo %d] nuevo mejor modelo (%d/210): %s.best%s", c, tot, modelPath, "\n")
		}
		fmt.Printf("[ciclo %d] CON biblio (%d lemas): %s total=%d/210\n", c, len(LemmaLib.Lemmas), metric, tot)
		// ablacion en el primer y ultimo ciclo: mismo net, SIN biblioteca
		if c == 1 || c == nCycles {
			saved := LemmaLib
			LemmaLib = nil
			m2, t2 := evalAll(net, lv, cap)
			LemmaLib = saved
			fmt.Printf("[ciclo %d] SIN biblio (ablacion): %s total=%d/210\n", c, m2, t2)
		}
		// 2) AUTOJUEGO + verificacion + extraccion de lemas
		var buf []Sample
		var hindsight []HindsightRecord
		solved, nuevos := 0, 0
		for i := 0; i < perRound; i++ {
			var a, b *Expr
			accepted := false
			for attempt := 0; attempt < 1000; attempt++ {
				a, b = genTrain(r, net.Level)
				if !fixedPairs[RenderPrefix(a)+"|"+RenderPrefix(b)] {
					accepted = true
					break
				}
			}
			if !accepted {
				fmt.Println("[cycle] cannot obtain train pair separate from fixed evaluation")
				return
			}
			res := guidedSolveBidirectional(net, a, b, cap, true)
			ok, sm := res.OK, res.Samples
			if res.Hindsight != nil {
				// Also exclude fixed evaluation pairs from hindsight labels.
				if !fixedPairs[res.Hindsight.Start+"|"+res.Hindsight.Target] {
					buf = append(buf, sm...)
					hindsight = append(hindsight, HindsightRecord{Seed: seed, Cycle: c, Problem: i, OriginalTarget: RenderPrefix(b), Cap: cap, Pair: *res.Hindsight})
				}
			}
			if ok {
				solved++
				buf = append(buf, sm...)
				if LemmaLib.Insert(a, b, len(sm)) { // juez verifica en Insert
					nuevos++
				}
			}
		}
		if !saveHindsightRecords(modelPath+".hindsight.json", hindsight) {
			fmt.Println("[cycle] cannot persist verified hindsight")
			return
		}
		// 3) ENTRENAMIENTO (imitar las acciones que resolvieron)
		for ep := 0; ep < 3; ep++ {
			r.Shuffle(len(buf), func(i, j int) { buf[i], buf[j] = buf[j], buf[i] })
			for _, s := range buf {
				net.TrainCE(s.Feats, s.Target, 0.003)
			}
		}
		net.Round++
		if float64(solved)/float64(perRound) >= 0.8 && net.Level < 8 {
			net.Level++
			nombre := fmt.Sprintf("%s.nivel%d", modelPath, net.Level)
			net.Save(nombre)
			fmt.Printf("[ciclo %d] nivel superado: checkpoint %s%s", c, nombre, "\n")
		}
		fmt.Printf("[ciclo %d] entreno: nivel=%d resueltos=%d/%d lemas_nuevos=%d total_lemas=%d\n",
			c, net.Level, solved, perRound, nuevos, len(LemmaLib.Lemmas))
		// Utility is trained only on train searches, never fixed metrics. Reindex
		// only after all samples have been consumed and no action indices survive.
		beforePrune := len(LemmaLib.Lemmas)
		LemmaLib.PruneUtility()
		fmt.Printf("[ciclo %d] utility prune=%d->%d hindsight_pairs=%d\n", c, beforePrune, len(LemmaLib.Lemmas), len(hindsight))
		// 4) PERSISTENCIA: todo versionado en disco tras cada ciclo
		net.Save(modelPath)
		LemmaLib.Save(libPath)
	}
	fmt.Println("[cycle] modelo y biblioteca guardados:", modelPath, libPath)
}

func CycleMain(args []string) {
	n, per, cap := 6, 150, 2000
	if len(args) > 0 {
		n, _ = strconv.Atoi(args[0])
	}
	if len(args) > 1 {
		per, _ = strconv.Atoi(args[1])
	}
	if len(args) > 2 {
		cap, _ = strconv.Atoi(args[2])
	}
	// Optional trailing lambda,N,activeCap; reject invalid values, no fallback.
	if len(args) > 3 {
		x, err := strconv.ParseFloat(args[3], 64)
		if err != nil || math.IsNaN(x) || math.IsInf(x, 0) || x < 0 {
			fmt.Println("invalid utility lambda")
			return
		}
		LemmaUtilityLambda = x
	}
	if len(args) > 4 {
		x, err := strconv.Atoi(args[4])
		if err != nil || x < 1 {
			fmt.Println("invalid utility N")
			return
		}
		LemmaUtilityMinProblems = x
	}
	if len(args) > 5 {
		x, err := strconv.Atoi(args[5])
		if err != nil || x < 64 || x > 128 {
			fmt.Println("utility active cap must be64..128")
			return
		}
		LemmaActiveCap = x
	}
	Cycle(n, per, cap, "model_math.json", "lemas.json", 1)
}

// Append-only logical record; atomic replacement avoids partial JSON writes.
// No evaluation calls reach this function.
type HindsightRecord struct {
	Seed           int64         `json:"seed"`
	Cycle          int           `json:"cycle"`
	Problem        int           `json:"problem"`
	OriginalTarget string        `json:"original_target"`
	Cap            int           `json:"cap"`
	Pair           HindsightPair `json:"pair"`
}

func saveHindsightRecords(path string, added []HindsightRecord) bool {
	if len(added) == 0 {
		return true
	}
	var records []HindsightRecord
	if data, err := os.ReadFile(path); err == nil {
		if json.Unmarshal(data, &records) != nil {
			return false
		}
	} else if !os.IsNotExist(err) {
		return false
	}
	records = append(records, added...)
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return false
	}
	if os.WriteFile(path+".tmp", data, 0600) != nil {
		return false
	}
	if os.Rename(path+".tmp", path) != nil {
		os.Remove(path + ".tmp")
		return false
	}
	return true
}

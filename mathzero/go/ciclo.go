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
	"fmt"
	"math/rand"
	"strconv"
)

func evalAll(net *MLP, lv map[int][]MState, cap int) (string, int) {
	s := ""
	tot := 0
	for k := 2; k <= 8; k++ {
		o, nd := EvalLevel(net, lv[k], cap)
		s += fmt.Sprintf("k%d:%d/30(%d) ", k, o, nd)
		tot += o
	}
	return s, tot
}

func Cycle(nCycles, perRound, cap int, modelPath, libPath string, seed int64) {
	// E11 aceptado (ver RESULTADOS_MEJORAS.md): motor con direccion
	// explicita + dedup de sucesores. El pipeline real lo usa por defecto.
	DirFix = true
	r := rand.New(rand.NewSource(seed))
	net := LoadMLP(modelPath)
	if net == nil {
		net = NewMLP(seed)
		fmt.Println("[cycle] red nueva")
	} else {
		fmt.Printf("[cycle] continuando desde %s (nivel %d, ronda %d)\n", modelPath, net.Level, net.Round)
	}
	full := LoadLibrary(libPath)
	// E16 (cierre de E9, 2026-10-17): en el ciego canonico (ab e9 seed=12345,
	// mismo net) vacio=539 > top-8=526 > cruda=520 /700. La biblioteca NO bate
	// al vacio: el pipeline shipped no usa lemas; el top-8 queda como ablacion.
	LemmaLib = nil
	fmt.Printf("[cycle] biblioteca: %d candidatos, 0 activos (E16: default sin lemas)\n", len(full.Lemmas))
	lv := map[int][]MState{}
	for k := 2; k <= 8; k++ {
		lv[k] = LoadLevels(k)
	}
	fmt.Println("[cycle] metrica fija: levels.json, cap", cap, "(resueltos y nodos medios por nivel)")
	bestTot := -1
	for c := 1; c <= nCycles; c++ {
		// 1) METRICA del estado actual (antes de aprender mas). Shipped: sin
		// biblioteca (E16).
		metric, tot := evalAll(net, lv, cap)
		if tot > bestTot {
			bestTot = tot
			net.Save(modelPath + ".best")
			fmt.Printf("[ciclo %d] nuevo mejor modelo (%d/210): %s.best%s", c, tot, modelPath, "\n")
		}
		fmt.Printf("[ciclo %d] shipped sin lemas: %s total=%d/210\n", c, metric, tot)
		// ablacion en el primer y ultimo ciclo: mismo net, top-8 por uso medido
		if c == 1 || c == nCycles {
			pod := full.Pruned(full.UsoMap(), 8)
			saved := LemmaLib
			LemmaLib = pod
			m2, t2 := evalAll(net, lv, cap)
			LemmaLib = saved
			fmt.Printf("[ciclo %d] ablacion top-8 (%d lemas): %s total=%d/210\n", c, len(pod.Lemmas), m2, t2)
		}
		// 2) AUTOJUEGO + verificacion + extraccion de lemas. E9: el autojuego
		// ve los candidatos CRUDOS para medir uso y anadir candidatos nuevos.
		LemmaLib = full
		var buf []Sample
		solved, nuevos := 0, 0
		for i := 0; i < perRound; i++ {
			a, b := genTrain(r, net.Level)
			ok, _, sm := GuidedSolve(net, a, b, cap)
			if ok {
				solved++
				buf = append(buf, sm...)
				if LemmaLib.Insert(a, b, len(sm)) { // juez verifica en Insert
					nuevos++
				}
			}
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
		// 4) PERSISTENCIA: todo versionado en disco tras cada ciclo
		net.Save(modelPath)
		full.Save(libPath)
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
	Cycle(n, per, cap, "model_math.json", "lemas.json", 1)
}

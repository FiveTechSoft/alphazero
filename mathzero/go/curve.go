// Curva de aprendizaje sobre el conjunto CIEGO (hito 16): un comando
//
//	./mathzero curve [ciclos] [problemas_por_ciclo] [cap]
//
// Imprime por ciclo: resueltos por nivel, nodos medios y longitud media
// de prueba (sobre resueltos), con ablaciones SIN red y SIN biblioteca
// en el primer y ultimo ciclo. El entrenamiento excluye cualquier
// problema del conjunto ciego (filtro por clave).
package main

import (
	"fmt"
	"math/rand"
	"strconv"
)

type blindEval struct {
	perLevel [9]int // resueltos por k
	total    int
	nodes    int
	proofs   int // longitud total de pruebas (solo resueltos)
	nSolved  int
}

func evalBlind(net *MLP, lv map[int][]MState, cap int) blindEval {
	var r blindEval
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
			}
		}
		r.nodes += nd / len(lv[k])
	}
	r.nodes /= 7
	return r
}

func CurveMain(args []string) {
	n, per, cap := 3, 60, 1000
	if len(args) > 0 {
		n, _ = strconv.Atoi(args[0])
	}
	if len(args) > 1 {
		per, _ = strconv.Atoi(args[1])
	}
	if len(args) > 2 {
		cap, _ = strconv.Atoi(args[2])
	}
	r := rand.New(rand.NewSource(1))
	net := LoadMLP("model_math.json")
	if net == nil {
		net = NewMLP(1)
		fmt.Println("[curve] red nueva")
	} else {
		fmt.Printf("[curve] continuando desde model_math.json (nivel %d, ronda %d)\n", net.Level, net.Round)
	}
	LemmaLib = LoadLibrary("lemas.json")
	blind := LoadBlindKeys("blind.json")
	fmt.Printf("[curve] conjunto ciego: %d problemas excluidos del entrenamiento; biblioteca: %d lemas\n",
		len(blind), len(LemmaLib.Lemmas))
	lv := map[int][]MState{}
	for k := 2; k <= 8; k++ {
		lv[k] = LoadProblems("blind.json", fmt.Sprint(k))
	}
	report := func(tag string, e blindEval) {
		pl := 0.0
		if e.nSolved > 0 {
			pl = float64(e.proofs) / float64(e.nSolved)
		}
		fmt.Printf("[curva %s] k2:%d k3:%d k4:%d k5:%d k6:%d k7:%d k8:%d  total=%d/700 nodos_med=%d prueba_med=%.1f\n",
			tag, e.perLevel[2], e.perLevel[3], e.perLevel[4], e.perLevel[5], e.perLevel[6], e.perLevel[7], e.perLevel[8],
			e.total, e.nodes, pl)
	}
	for c := 1; c <= n; c++ {
		report(fmt.Sprintf("%d CON todo (%d lemas)", c, len(LemmaLib.Lemmas)), evalBlind(net, lv, cap))
		if c == 1 || c == n {
			report(fmt.Sprintf("%d SIN red", c), evalBlind(nil, lv, cap))
			saved := LemmaLib
			LemmaLib = nil
			report(fmt.Sprintf("%d SIN biblio", c), evalBlind(net, lv, cap))
			LemmaLib = saved
		}
		// entrenamiento del ciclo (excluyendo el ciego por clave)
		var buf []Sample
		solved, nuevos := 0, 0
		for i := 0; i < per; i++ {
			a, b := genTrain(r, net.Level)
			if blind[pairKey(a, b)] {
				continue
			}
			ok, _, sm := GuidedSolve(net, a, b, cap)
			if ok {
				solved++
				buf = append(buf, sm...)
				if LemmaLib.Insert(a, b, len(sm)) {
					nuevos++
				}
			}
		}
		for ep := 0; ep < 3; ep++ {
			r.Shuffle(len(buf), func(i, j int) { buf[i], buf[j] = buf[j], buf[i] })
			for _, s := range buf {
				net.TrainCE(s.Feats, s.Target, 0.003)
			}
		}
		net.Round++
		if float64(solved)/float64(per) >= 0.8 && net.Level < 8 {
			net.Level++
		}
		fmt.Printf("[curva %d] entreno: nivel=%d resueltos=%d/%d lemas_nuevos=%d\n", c, net.Level, solved, per, nuevos)
		net.Save("model_math.json")
		LemmaLib.Save("lemas.json")
	}
	fmt.Println("[curve] modelo y biblioteca guardados")
}

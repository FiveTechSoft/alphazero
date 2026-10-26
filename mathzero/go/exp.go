// A/B de experimentos (E2 online, E4 replay priorizado, E8 transferencia,
// E9 poda de lemas por uso, E10 datos sinteticos neg-neg/add0).
// Un comando, dos brazos desde la MISMA semilla (init identico, mismo stream
// de problemas), evaluados en el mismo conjunto ciego con el mismo cap:
//
//	./mathzero ab <e2|e4|e8> <semilla> [rondas] [por_ronda] [cap]
//
// Sin biblioteca en los brazos (medido: los lemas crudos hunden la busqueda).
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

type ExpCfg struct {
	Online   bool // E2: entrena tras cada problema, sin acumular ronda
	Replay   bool // E4: replay priorizado por longitud de prueba
	MaxLevel int  // E8: techo de nivel de entrenamiento (0 = sin techo)
	SynthNeg bool // E10: 25% de objetivos envueltos con neg-neg/add0
	DirFix   bool // E11: motor con direccion explicita + dedup de sucesores
	Accept   bool // E13: puerta de aceptacion por set fijo (levels.json)
}

// wrapCleanup envuelve subexpresiones al azar con las formas que la red
// no sabe limpiar (medido en evalfail): doble negacion y sumar cero.
// El objetivo canonico no cambia: la prueba EXIGE negneg/add0.
func wrapCleanup(r *rand.Rand, e *Expr) *Expr {
	for i := 0; i < 1+r.Intn(3); i++ {
		ps := SubPositions(e)
		if len(ps) == 0 {
			break
		}
		p := ps[r.Intn(len(ps))]
		sub := Get(e, p)
		if r.Float64() < 0.5 {
			e = Replace(e, p, ENeg(ENeg(sub)))
		} else {
			e = Replace(e, p, EAdd(sub, C(0)))
		}
	}
	return e
}

type wsample struct {
	s Sample
	w float64
}

func trainArm(cfg ExpCfg, seed int64, rounds, per, cap int, blind map[string]bool) *MLP {
	r := rand.New(rand.NewSource(seed))
	net := NewMLP(seed)
	prevDirFix := DirFix
	DirFix = cfg.DirFix
	defer func() { DirFix = prevDirFix }()
	var lvFix map[int][]MState
	var prevPath string
	if cfg.Accept {
		lvFix = map[int][]MState{}
		for k := 2; k <= 8; k++ {
			lvFix[k] = LoadLevels(k)
		}
		if f, err := os.CreateTemp("", "e13prev*.json"); err == nil {
			prevPath = f.Name()
			f.Close()
			defer os.Remove(prevPath)
		}
	}
	var replay []wsample
	for rd := 1; rd <= rounds; rd++ {
		var buf []Sample
		solved := 0
		for i := 0; i < per; i++ {
			lvl := net.Level
			if cfg.MaxLevel > 0 && lvl > cfg.MaxLevel {
				lvl = cfg.MaxLevel
			}
			a, b := genTrain(r, lvl)
			if blind[pairKey(a, b)] {
				continue
			}
			if cfg.SynthNeg && r.Float64() < 0.25 {
				b = wrapCleanup(r, b)
			}
			ok, _, sm := GuidedSolve(net, a, b, cap)
			if !ok {
				continue
			}
			solved++
			switch {
			case cfg.Online:
				for ep := 0; ep < 3; ep++ {
					for _, s := range sm {
						net.TrainCE(s.Feats, s.Target, 0.003)
					}
				}
			case cfg.Replay:
				for _, s := range sm {
					replay = append(replay, wsample{s, float64(len(sm))})
				}
			default:
				buf = append(buf, sm...)
			}
		}
		if cfg.Accept && prevPath != "" {
			net.Save(prevPath) // copia del estado previo a la ronda
		}
		if cfg.Replay && len(replay) > 0 {
			total := 0.0
			for _, x := range replay {
				total += x.w
			}
			ndraws := len(buf) + per*3 // presupuesto comparable al baseline
			for ep := 0; ep < 3; ep++ {
				for d := 0; d < ndraws; d++ {
					u := r.Float64() * total
					acc := 0.0
					for _, x := range replay {
						acc += x.w
						if acc >= u {
							net.TrainCE(x.s.Feats, x.s.Target, 0.003)
							break
						}
					}
				}
			}
			if len(replay) > 20000 {
				replay = replay[len(replay)-20000:]
			}
		} else if !cfg.Online && len(buf) > 0 {
			for ep := 0; ep < 3; ep++ {
				r.Shuffle(len(buf), func(i, j int) { buf[i], buf[j] = buf[j], buf[i] })
				for _, s := range buf {
					net.TrainCE(s.Feats, s.Target, 0.003)
				}
			}
		}
		if cfg.Accept && prevPath != "" {
			prev := LoadMLP(prevPath)
			_, tPrev := evalAll(prev, lvFix, cap)
			_, tNew := evalAll(net, lvFix, cap)
			if tNew < tPrev {
				net = prev
				fmt.Printf("[ab e13 ronda %d] set fijo: nuevo %d/210 < previo %d/210 -> RECHAZADA%s", rd, tNew, tPrev, "\n")
			} else {
				fmt.Printf("[ab e13 ronda %d] set fijo: nuevo %d/210 >= previo %d/210 -> aceptada%s", rd, tNew, tPrev, "\n")
			}
		}
		net.Round++
		maxL := 8
		if cfg.MaxLevel > 0 {
			maxL = cfg.MaxLevel
		}
		if float64(solved)/float64(per) >= 0.8 && net.Level < maxL {
			net.Level++
		}
	}
	return net
}

func abReport(tag string, e blindEval) {
	fmt.Printf("[ab %s] k2:%d k3:%d k4:%d k5:%d k6:%d k7:%d k8:%d total=%d/700 k6-8=%d nodos_med=%d\n",
		tag, e.perLevel[2], e.perLevel[3], e.perLevel[4], e.perLevel[5],
		e.perLevel[6], e.perLevel[7], e.perLevel[8], e.total,
		e.perLevel[6]+e.perLevel[7]+e.perLevel[8], e.nodes)
}

type ABArmJSON struct {
	PorNivel map[string]int `json:"por_nivel"`
	Total    int            `json:"total"`
	K68      int            `json:"k6_8"`
	NodosMed int            `json:"nodos_med"`
}

type ABResultJSON struct {
	Experimento string    `json:"experimento"`
	Semilla     int64     `json:"semilla"`
	Rondas      int       `json:"rondas"`
	PorRonda    int       `json:"por_ronda"`
	Cap         int       `json:"cap"`
	Baseline    ABArmJSON `json:"baseline"`
	Variante    ABArmJSON `json:"variante"`
}

func armJSON(e blindEval) ABArmJSON {
	pl := map[string]int{}
	for k := 2; k <= 8; k++ {
		pl[fmt.Sprintf("k%d", k)] = e.perLevel[k]
	}
	return ABArmJSON{PorNivel: pl, Total: e.total,
		K68: e.perLevel[6] + e.perLevel[7] + e.perLevel[8], NodosMed: e.nodes}
}

func ABMain(args []string) {
	wantJSON := false
	var pos []string
	for _, a := range args {
		if a == "-json" {
			wantJSON = true
		} else {
			pos = append(pos, a)
		}
	}
	args = pos
	if len(args) < 2 || len(args) > 5 {
		fmt.Println("uso: ./mathzero ab <e2|e4|e8|e9|e10|e11|e13> <semilla> [rondas] [por_ronda] [cap] [-json]")
		return
	}
	exp := args[0]
	seed, ok := parseInt64Arg("semilla", args[1])
	if !ok {
		return
	}
	rounds, per, cap := 8, 60, 1000
	if len(args) > 2 {
		if rounds, ok = parseIntArg("rondas", args[2], 1); !ok {
			return
		}
	}
	if len(args) > 3 {
		if per, ok = parseIntArg("por_ronda", args[3], 1); !ok {
			return
		}
	}
	if len(args) > 4 {
		if cap, ok = parseIntArg("cap", args[4], 1); !ok {
			return
		}
	}
	LemmaLib = nil
	blind := LoadBlindKeys("blind.json")
	lv := map[int][]MState{}
	for k := 2; k <= 8; k++ {
		lv[k] = LoadProblems("blind.json", fmt.Sprint(k))
	}
	var vari ExpCfg
	switch exp {
	case "e2":
		vari = ExpCfg{Online: true}
	case "e4":
		vari = ExpCfg{Replay: true}
	case "e8":
		vari = ExpCfg{MaxLevel: 5}
	case "e10":
		vari = ExpCfg{SynthNeg: true}
	case "e11":
		vari = ExpCfg{DirFix: true}
	case "e13":
		vari = ExpCfg{Accept: true}
	case "e9":
		// Poda de lemas-accion por uso medido: UN entrenamiento con la
		// biblioteca cruda midiendo uso; mismo modelo evaluado con la
		// biblioteca cruda (baseline) y con la podada top-8 (variante).
		lib := LoadLibrary("lemas.json")
		fmt.Printf("[ab e9 seed=%d] rondas=%d por_ronda=%d cap=%d, biblioteca=%d lemas\n",
			seed, rounds, per, cap, len(lib.Lemmas))
		LemmaLib = lib
		LemmaUso = map[int]int{}
		net9 := trainArm(ExpCfg{}, seed, rounds, per, cap, blind)
		podada := lib.Pruned(LemmaUso, 8)
		fmt.Printf("[ab e9 seed=%d] lemas con uso medido: %d de %d; podada queda en %d\n",
			seed, len(LemmaUso), len(lib.Lemmas), len(podada.Lemmas))
		ebase9 := evalBlind(net9, lv, cap)
		abReport(fmt.Sprintf("e9 seed=%d baseline (cruda %d lemas)", seed, len(lib.Lemmas)), ebase9)
		LemmaLib = podada
		evari9 := evalBlind(net9, lv, cap)
		abReport(fmt.Sprintf("e9 seed=%d variante (podada %d lemas)", seed, len(podada.Lemmas)), evari9)
		// E16 (cierre de E9, 2026-10-17): brazo VACIO, mismo net/ciego, para
		// medir en el ciego canonico si la top-8 bate a no usar biblioteca.
		LemmaLib = nil
		eempty9 := evalBlind(net9, lv, cap)
		abReport(fmt.Sprintf("e9 seed=%d vacio (sin lemas)", seed), eempty9)
		LemmaLib, LemmaUso = nil, nil
		if wantJSON {
			out := ABResultJSON{
				Experimento: exp, Semilla: seed, Rondas: rounds, PorRonda: per, Cap: cap,
				Baseline: armJSON(ebase9), Variante: armJSON(evari9),
			}
			name := fmt.Sprintf("resultados_ab_%s_seed%d.json", exp, seed)
			data, _ := json.MarshalIndent(out, "", "  ")
			if err := writeFileAtomic(name, data); err != nil {
				fmt.Println("[ab] error escribiendo", name, ":", err)
			} else {
				fmt.Println("[ab] resultados escritos en", name)
			}
		}
		return
	default:
		fmt.Println("experimento desconocido:", exp)
		return
	}
	fmt.Printf("[ab %s seed=%d] rondas=%d por_ronda=%d cap=%d (sin biblioteca, mismo ciego)\n", exp, seed, rounds, per, cap)
	base := trainArm(ExpCfg{}, seed, rounds, per, cap, blind)
	ebase := evalBlind(base, lv, cap)
	abReport(fmt.Sprintf("%s seed=%d baseline", exp, seed), ebase)
	vnet := trainArm(vari, seed, rounds, per, cap, blind)
	DirFix = vari.DirFix // evaluar la variante con SU motor
	evari := evalBlind(vnet, lv, cap)
	DirFix = false
	abReport(fmt.Sprintf("%s seed=%d variante", exp, seed), evari)
	if wantJSON {
		out := ABResultJSON{
			Experimento: exp, Semilla: seed, Rondas: rounds, PorRonda: per, Cap: cap,
			Baseline: armJSON(ebase), Variante: armJSON(evari),
		}
		name := fmt.Sprintf("resultados_ab_%s_seed%d.json", exp, seed)
		data, _ := json.MarshalIndent(out, "", "  ")
		if err := writeFileAtomic(name, data); err != nil {
			fmt.Println("[ab] error escribiendo", name, ":", err)
		} else {
			fmt.Println("[ab] resultados escritos en", name)
		}
	}
}

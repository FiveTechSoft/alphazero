// Busqueda guiada por la red (mejor-primero por log-prob acumulada) +
// expert iteration con curriculo. Mismo cap de nodos que el baseline BFS.
package main

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
)

func size(e *Expr) int {
	if e.Op == "const" || e.Op == "var" {
		return 1
	}
	if e.Op == "neg" {
		return 1 + size(e.A)
	}
	return 1 + size(e.A) + size(e.B)
}
func count(e *Expr, op string) int {
	if e == nil {
		return 0
	}
	c := 0
	if e.Op == op {
		c = 1
	}
	return c + count(e.A, op) + count(e.B, op)
}

// features de (estado, accion): 14 one-hot regla + 10 numericas
func Feats(cur, target *Expr, a MAction) []float64 {
	f := make([]float64, NFeat)
	if a.Rule >= 0 {
		f[a.Rule] = 1
	}
	nxt := applyAction(cur, a)
	sb, sa, st := size(cur), size(nxt), size(target)
	f[14] = float64(len(a.Pos)) / 5
	f[15] = float64(sb-st) / 10
	f[16] = float64(sa-st) / 10
	f[17] = (math.Abs(float64(sa-st)) - math.Abs(float64(sb-st))) / 5
	f[18] = float64(count(nxt, "mul")-count(target, "mul")) / 5
	f[19] = float64(count(nxt, "add")-count(target, "add")) / 5
	f[20] = float64(count(nxt, "neg")-count(target, "neg")) / 3
	f[21] = (math.Abs(float64(count(nxt, "mul")-count(target, "mul"))) - math.Abs(float64(count(cur, "mul")-count(target, "mul")))) / 3
	f[22] = (math.Abs(float64(count(nxt, "add")-count(target, "add"))) - math.Abs(float64(count(cur, "add")-count(target, "add")))) / 3
	if Key(nxt) == Key(target) {
		f[23] = 1
	}
	return f
}

type item struct {
	cost   float64
	e      *Expr
	parent *item
	acts   []MAction // acciones del padre (para reconstruir)
	aidx   int
	feats  [][]float64
}
type pq []*item

func (p pq) Len() int            { return len(p) }
func (p pq) Less(i, j int) bool  { return p[i].cost < p[j].cost }
func (p pq) Swap(i, j int)       { p[i], p[j] = p[j], p[i] }
func (p *pq) Push(x interface{}) { *p = append(*p, x.(*item)) }
func (p *pq) Pop() interface{}   { o := *p; n := len(o); x := o[n-1]; *p = o[:n-1]; return x }

type Sample struct {
	Feats  [][]float64
	Target int
}

// GuidedSolve: net==nil -> prior uniforme. Devuelve ok, nodos, muestras del camino.
func GuidedSolve(net *MLP, start, target *Expr, cap int) (bool, int, []Sample) {
	ClearKeyCache()
	tk := Key(target)
	seen := map[string]bool{Key(start): true}
	h := &pq{}
	heap.Push(h, &item{e: start})
	nodes := 0
	for h.Len() > 0 && nodes < cap {
		it := heap.Pop(h).(*item)
		nodes++
		if Key(it.e) == tk {
			var out []Sample
			for x := it; x.parent != nil; x = x.parent {
				out = append(out, Sample{x.parent.feats, x.aidx})
			}
			if LemmaLib != nil && LemmaUso != nil {
				for x := it; x.parent != nil; x = x.parent {
					acts := MathActions(x.parent.e)
					if x.aidx >= 0 && x.aidx < len(acts) && acts[x.aidx].Lemma >= 0 {
						LemmaUso[acts[x.aidx].Lemma]++
					}
				}
			}
			return true, nodes, out
		}
		acts := MathActions(it.e)
		if len(acts) == 0 {
			continue
		}
		fs := make([][]float64, len(acts))
		sc := make([]float64, len(acts))
		mx := -1e18
		for i, a := range acts {
			fs[i] = Feats(it.e, target, a)
			if net != nil {
				sc[i] = net.ScoreClean(fs[i])
			}
			if sc[i] > mx {
				mx = sc[i]
			}
		}
		z := 0.0
		for _, s := range sc {
			z += math.Exp(s - mx)
		}
		it.feats = fs
		for i, a := range acts {
			nx := applyAction(it.e, a)
			k := Key(nx)
			if seen[k] {
				continue
			}
			seen[k] = true
			lp := sc[i] - mx - math.Log(z)
			heap.Push(h, &item{cost: it.cost - lp, e: nx, parent: it, aidx: i})
		}
	}
	return false, nodes, nil
}

func EvalLevel(net *MLP, probs []MState, cap int) (int, int) {
	ok, nodes := 0, 0
	for _, p := range probs {
		s, n, _ := GuidedSolve(net, p.Cur, p.Target, cap)
		if s {
			ok++
		}
		nodes += n
	}
	return ok, nodes / len(probs)
}

// generador de problemas nuevos para entrenar (no los de evaluacion)
func genTrain(r *rand.Rand, k int) (*Expr, *Expr) {
	var rnd func(d int) *Expr
	rnd = func(d int) *Expr {
		if d == 0 || r.Float64() < 0.3 {
			if r.Float64() < 0.6 {
				return V(string(rune('a' + r.Intn(3))))
			}
			return C(r.Intn(8) - 3)
		}
		switch r.Intn(3) {
		case 0:
			return ENeg(rnd(d - 1))
		case 1:
			return EAdd(rnd(d-1), rnd(d-1))
		}
		return EMul(rnd(d-1), rnd(d-1))
	}
	a := rnd(2 + r.Intn(2))
	cur := a
	for i := 0; i < k; i++ {
		acts := MathActions(cur)
		if len(acts) == 0 {
			break
		}
		cur = applyAction(cur, acts[r.Intn(len(acts))])
	}
	return a, cur
}

func LoadLevels(k int) []MState { return LoadProblems("levels.json", fmt.Sprint(k)) }

func Curriculum(rounds, perRound, cap int, modelPath string, seed int64) {
	r := rand.New(rand.NewSource(seed))
	net := LoadMLP(modelPath)
	if net == nil {
		net = NewMLP(seed)
		fmt.Println("[learn] red nueva")
	} else {
		fmt.Printf("[learn] continuando desde %s (nivel %d, ronda %d)\n", modelPath, net.Level, net.Round)
	}
	lv := map[int][]MState{}
	for k := 2; k <= 8; k++ {
		lv[k] = LoadLevels(k)
	}
	report := func(tag string) {
		fmt.Printf("[eval %s] cap=%d  ", tag, cap)
		for k := 2; k <= 8; k++ {
			o, _ := EvalLevel(net, lv[k], cap)
			fmt.Printf("k%d:%d/30 ", k, o)
		}
		fmt.Println()
	}
	fmt.Print("[baseline BFS] ")
	for k := 2; k <= 8; k++ {
		o := 0
		for _, p := range lv[k] {
			if s, _ := BFSSolve(MathRing{}, p, cap); s {
				o++
			}
		}
		fmt.Printf("k%d:%d/30 ", k, o)
	}
	fmt.Println()
	report("red sin entrenar")
	for rd := 1; rd <= rounds; rd++ {
		var buf []Sample
		solved := 0
		for i := 0; i < perRound; i++ {
			a, b := genTrain(r, net.Level)
			ok, _, sm := GuidedSolve(net, a, b, cap)
			if ok {
				solved++
				buf = append(buf, sm...)
			}
		}
		loss := 0.0
		for ep := 0; ep < 3; ep++ {
			r.Shuffle(len(buf), func(i, j int) { buf[i], buf[j] = buf[j], buf[i] })
			for _, s := range buf {
				loss += net.TrainCE(s.Feats, s.Target, 0.003)
			}
		}
		net.Round++
		rate := float64(solved) / float64(perRound)
		if rate >= 0.8 && net.Level < 8 {
			net.Level++
		}
		fmt.Printf("[ronda %d] nivel_entreno=%d resueltos %d/%d muestras=%d\n", rd, net.Level, solved, perRound, len(buf))
		if rd%3 == 0 || rd == rounds {
			report(fmt.Sprintf("ronda %d", rd))
		}
	}
	net.Save(modelPath)
	fmt.Println("[learn] modelo guardado en", modelPath)
}

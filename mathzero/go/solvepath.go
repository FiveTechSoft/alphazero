// SolvePath: misma busqueda mejor-primero guiada por la red que
// GuidedSolve, pero devuelve el camino de expresiones (para mostrarlo).
package main

import (
	"container/heap"
	"math"
)

type PathStep struct {
	From, To *Expr
	Rule     string
}

type pitem struct {
	cost   float64
	e      *Expr
	parent *pitem
	act    MAction
}

type ppq []*pitem

func (p ppq) Len() int            { return len(p) }
func (p ppq) Less(i, j int) bool  { return p[i].cost < p[j].cost }
func (p ppq) Swap(i, j int)       { p[i], p[j] = p[j], p[i] }
func (p *ppq) Push(x interface{}) { *p = append(*p, x.(*pitem)) }
func (p *ppq) Pop() interface{} {
	o := *p
	n := len(o)
	x := o[n-1]
	*p = o[:n-1]
	return x
}

func SolvePath(net *MLP, start, target *Expr, cap int) (bool, int, []PathStep) {
	ClearKeyCache()
	tk := Key(target)
	seen := map[string]bool{Key(start): true}
	h := &ppq{}
	heap.Push(h, &pitem{e: start})
	nodes := 0
	for h.Len() > 0 && nodes < cap {
		it := heap.Pop(h).(*pitem)
		nodes++
		if Key(it.e) == tk {
			var rev []PathStep
			for x := it; x.parent != nil; x = x.parent {
				rev = append(rev, PathStep{From: x.parent.e, To: x.e, Rule: RuleName(x.act)})
			}
			for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
				rev[i], rev[j] = rev[j], rev[i]
			}
			return true, nodes, rev
		}
		acts := MathActions(it.e)
		if len(acts) == 0 {
			continue
		}
		sc := make([]float64, len(acts))
		mx := -1e18
		for i, a := range acts {
			if net != nil {
				sc[i] = net.ScoreClean(Feats(it.e, target, a))
			}
			if sc[i] > mx {
				mx = sc[i]
			}
		}
		z := 0.0
		for _, s := range sc {
			z += math.Exp(s - mx)
		}
		for i, a := range acts {
			nx := applyAction(it.e, a)
			k := Key(nx)
			if seen[k] {
				continue
			}
			seen[k] = true
			lp := sc[i] - mx - math.Log(z)
			heap.Push(h, &pitem{cost: it.cost - lp, e: nx, parent: it, act: a})
		}
	}
	return false, nodes, nil
}

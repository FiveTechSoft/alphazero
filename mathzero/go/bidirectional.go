package main

import (
	"container/heap"
	"math"
)

// Training-only side effects are opt-in. Evaluation/default calls never change
// utility counters and never generate hindsight samples or records.
var SearchTraining bool

type HindsightPair struct {
	Start  string `json:"start"`
	Target string `json:"target"`
	Steps  int    `json:"steps"`
}

type bidirNode struct {
	e      *Expr
	parent *bidirNode
	cost   float64
	depth  int
	lemma  int
	serial int
}
type bidirQueue []*bidirNode

func (q bidirQueue) Len() int { return len(q) }
func (q bidirQueue) Less(i, j int) bool {
	if q[i].cost != q[j].cost {
		return q[i].cost < q[j].cost
	}
	return q[i].serial < q[j].serial
}
func (q bidirQueue) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *bidirQueue) Push(x interface{}) { *q = append(*q, x.(*bidirNode)) }
func (q *bidirQueue) Pop() interface{}   { a := *q; x := a[len(a)-1]; *q = a[:len(a)-1]; return x }

type bidirResult struct {
	OK        bool
	Nodes     int
	Samples   []Sample
	Hindsight *HindsightPair
}

// exactAction accepts only an enumerated concrete edge to exactly the desired
// AST. Key is AC-normalized and is deliberately NOT used for intersections.
func exactAction(from, to *Expr) ([]MAction, int) {
	acts := MathActions(from)
	want := RenderPrefix(to)
	for i, a := range acts {
		if RenderPrefix(applyAction(from, a)) == want {
			return acts, i
		}
	}
	return acts, -1
}
func verifiedEdge(a, b *Expr) bool { return canonEq(Canon(a), Canon(b)) }

func nodesFromRoot(n *bidirNode) []*bidirNode {
	var path []*bidirNode
	for ; n != nil; n = n.parent {
		path = append(path, n)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// Rebuild labels/features with the FINAL target, including every reversed
// b-side edge. Never reuse forward labels whose target was the opposite root.
func samplesForPath(path []*bidirNode, target *Expr) ([]Sample, []int, bool) {
	var samples []Sample
	var lemmas []int
	for i := 0; i+1 < len(path); i++ {
		from, to := path[i].e, path[i+1].e
		acts, idx := exactAction(from, to)
		if idx < 0 || !verifiedEdge(from, to) {
			return nil, nil, false
		}
		fs := make([][]float64, len(acts))
		for j, a := range acts {
			fs[j] = Feats(from, target, a)
		}
		samples = append(samples, Sample{Feats: fs, Target: idx})
		lemmas = append(lemmas, acts[idx].Lemma)
	}
	return samples, lemmas, true
}

// API preserved. Failed search returns hindsight samples ONLY in training mode;
// the boolean still means the ORIGINAL target was solved, not the hindsight one.
func GuidedSolve(net *MLP, start, target *Expr, cap int) (bool, int, []Sample) {
	r := guidedSolveBidirectional(net, start, target, cap, SearchTraining)
	return r.OK, r.Nodes, r.Samples
}

func guidedSolveBidirectional(net *MLP, start, target *Expr, cap int, training bool) bidirResult {
	ClearKeyCache()
	result := bidirResult{}
	if cap <= 0 {
		return result
	}
	if RenderPrefix(start) == RenderPrefix(target) {
		result.OK = true
		return result
	}
	roots := [2]*bidirNode{{e: start, lemma: -1}, {e: target, lemma: -1}}
	queues := [2]*bidirQueue{&bidirQueue{}, &bidirQueue{}}
	seen := [2]map[string]*bidirNode{{RenderPrefix(start): roots[0]}, {RenderPrefix(target): roots[1]}}
	for side := 0; side < 2; side++ {
		heap.Push(queues[side], roots[side])
	}
	budgets := [2]int{cap / 2, cap - cap/2}
	bestCost := [2]map[string]float64{{RenderPrefix(start): 0}, {RenderPrefix(target): 0}}
	used := [2]int{}
	serial := 0
	var generated []*bidirNode
	var winner map[*bidirNode]bool
	var winningLemmas []int
	if training && LemmaLib != nil {
		for _, l := range LemmaLib.Lemmas {
			l.TrainingProblems++
		}
	}
	finish := func() bidirResult {
		if training && LemmaLib != nil {
			for _, n := range generated {
				if n.lemma >= 0 && !winner[n] {
					LemmaLib.Lemmas[n.lemma].FailedGeneratedNodes++
				}
			}
			// Hindsight is training data, NOT a win on the original target.
			// Its edges do not get utility winner credit on a failed problem.
			if result.OK {
				for _, li := range winningLemmas {
					if li >= 0 {
						l := LemmaLib.Lemmas[li]
						l.WinningApplications++
						l.Uses++
						if LemmaUso != nil {
							LemmaUso[li]++
						}
					}
				}
			}
		}
		return result
	}
	var best *bidirNode
	bestKey := ""
	for {
		progress := false
		for side := 0; side < 2; side++ {
			if used[side] >= budgets[side] || queues[side].Len() == 0 {
				continue
			}
			progress = true
			it := heap.Pop(queues[side]).(*bidirNode)
			if searchStale(bestCost[side], RenderPrefix(it.e), it.cost) {
				continue
			}
			used[side]++
			result.Nodes++
			if side == 0 && it.depth >= 3 && (best == nil || size(it.e) < size(best.e) || (size(it.e) == size(best.e) && RenderPrefix(it.e) < bestKey)) {
				best = it
				bestKey = RenderPrefix(it.e)
			}
			acts := MathActions(it.e)
			goal := target
			if side == 1 {
				goal = start
			}
			valid := make([]int, 0, len(acts))
			scores := make([]float64, len(acts))
			mx := math.Inf(-1)
			for i, a := range acts {
				if training && LemmaLib != nil && a.Lemma >= 0 {
					LemmaLib.Lemmas[a.Lemma].MatchingAttempts++
				}
				nx := applyAction(it.e, a)
				if !verifiedEdge(it.e, nx) {
					continue
				}
				if side == 1 && !backwardEdgeAllowed(it.e, nx, a) {
					continue
				}

				valid = append(valid, i)
				if net != nil {
					scores[i] = net.ScoreClean(Feats(it.e, goal, a))
				}
				if scores[i] > mx {
					mx = scores[i]
				}
			}
			z := 0.0
			for _, i := range valid {
				z += math.Exp(scores[i] - mx)
			}
			logZ := math.Log(z)
			for _, i := range valid {
				a := acts[i]
				nx := applyAction(it.e, a)
				key := RenderPrefix(nx)
				cost := it.cost - (scores[i] - mx - logZ)
				if !searchRelax(bestCost[side], key, cost) {
					continue
				}
				serial++
				n := &bidirNode{e: nx, parent: it, depth: it.depth + 1, lemma: a.Lemma, serial: serial, cost: cost}
				seen[side][key] = n
				generated = append(generated, n)
				heap.Push(queues[side], n)
				if other := seen[1-side][key]; other != nil {
					left, right := n, other
					if side == 1 {
						left, right = other, n
					}
					path := nodesFromRoot(left)
					for x := right.parent; x != nil; x = x.parent {
						path = append(path, x)
					}
					sm, lm, ok := samplesForPath(path, target)
					if !ok {
						continue
					}
					result.OK = true
					result.Samples = sm
					winningLemmas = lm
					winner = map[*bidirNode]bool{}
					idx := left.depth - 1
					for x := left; x.parent != nil; x = x.parent {
						if x.lemma >= 0 && winningLemmas[idx] == x.lemma {
							winner[x] = true
						}
						idx--
					}
					idx = left.depth
					for x := right; x.parent != nil; x = x.parent {
						if x.lemma >= 0 && winningLemmas[idx] == x.lemma {
							winner[x] = true
						}
						idx++
					}
					return finish()
				}
			}
		}
		if !progress {
			break
		}
	}
	// No quota transfer when one frontier exhausts: each side has at most its
	// half. Hindsight requires TOTAL budget exhaustion, not an early dead end.
	if training && result.Nodes == cap && best != nil && verifiedEdge(start, best.e) {
		path := nodesFromRoot(best)
		sm, _, ok := samplesForPath(path, best.e)
		if ok && len(sm) >= 3 {
			result.Samples = sm
			result.Hindsight = &HindsightPair{Start: RenderPrefix(start), Target: RenderPrefix(best.e), Steps: len(sm)}
		}
	}
	return finish()
}

// Known information-losing rules have no finite concrete reverse enumerator.
func backwardEdgeAllowed(parent, child *Expr, a MAction) bool {
	if a.Rule == 2 || a.Rule == 10 || a.Rule == 12 || a.Rule == 13 {
		return false
	}
	_, inverse := exactAction(child, parent)
	return inverse >= 0 && verifiedEdge(parent, child)
}

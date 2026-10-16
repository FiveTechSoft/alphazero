package main

import "sort"

// Defaults are configurable through the documented CycleMain trailing arguments.
const DefaultLemmaUtilityLambda = 0.1
const DefaultLemmaUtilityMinProblems = 32
const DefaultLemmaActiveCap = 96

var LemmaUtilityLambda = DefaultLemmaUtilityLambda
var LemmaUtilityMinProblems = DefaultLemmaUtilityMinProblems
var LemmaActiveCap = DefaultLemmaActiveCap

// Saving is the original verified path length compressed into one lemma edge.
// Saving-1 is a path-length proxy, not measured causal search-node savings.
func (l *Lemma) Utility(lambda float64) float64 {
	return float64(l.WinningApplications)*float64(l.Saving-1) - lambda*float64(l.FailedGeneratedNodes)
}

// Prune only between complete training problems/cycles, never during a search
// or while sample/action indices are still being consumed.
func (L *Library) PruneUtility() {
	keep := make([]*Lemma, 0, len(L.Lemmas))
	for _, l := range L.Lemmas {
		if l.TrainingProblems >= L.UtilityMinProblems && l.Utility(L.UtilityLambda) <= 0 {
			continue
		}
		keep = append(keep, l)
	}
	sort.SliceStable(keep, func(i, j int) bool {
		a, b := keep[i], keep[j]
		if a.Utility(L.UtilityLambda) != b.Utility(L.UtilityLambda) {
			return a.Utility(L.UtilityLambda) > b.Utility(L.UtilityLambda)
		}
		if a.FromKey != b.FromKey {
			return a.FromKey < b.FromKey
		}
		return a.ToKey < b.ToKey
	})
	if L.ActiveCap > 0 && len(keep) > L.ActiveCap {
		keep = keep[:L.ActiveCap]
	}
	L.Lemmas = keep
	L.ByKey = map[string][]int{}
	for i, l := range keep {
		L.ByKey[l.FromKey] = append(L.ByKey[l.FromKey], i)
		L.ByKey[l.ToKey] = append(L.ByKey[l.ToKey], i)
	}
}

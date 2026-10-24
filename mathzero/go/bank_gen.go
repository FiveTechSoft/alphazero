package main

// P2 (new code, mathzero-next). Seeded bank generator with a declared rule,
// exclusion of the fixed banks, and content digests. Uses the base genTrain
// generator; it is not the Reduced8 bank.

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"sort"
)

// BankRule (declared): for each level k in Levels, draw genTrain(r, k) from one
// rand source seeded with Seed, in level order, rejecting a pair that is in
// levels.json, blind.json, or already drawn, until PerLevel pairs exist.
type BankRule struct {
	Seed     int64
	PerLevel int
	Levels   []int
}

type BankCase struct {
	EvalCase
	Level int
}

func fixedBankKeys() map[string]bool {
	used := map[string]bool{}
	for _, f := range []string{"levels.json", "blind.json"} {
		for k := 2; k <= 8; k++ {
			for _, p := range LoadProblems(f, fmt.Sprint(k)) {
				used[pairKey(p.Cur, p.Target)] = true
			}
		}
	}
	return used
}

// GenBank is deterministic in (rule, content of levels.json and blind.json).
func GenBank(rule BankRule) []BankCase {
	r := rand.New(rand.NewSource(rule.Seed))
	used := fixedBankKeys()
	var out []BankCase
	for _, k := range rule.Levels {
		n := 0
		for tries := 0; n < rule.PerLevel && tries < 200*rule.PerLevel+1000; tries++ {
			a, b := genTrain(r, k)
			pk := pairKey(a, b)
			if used[pk] || RenderPrefix(a) == RenderPrefix(b) {
				continue
			}
			used[pk] = true
			out = append(out, BankCase{EvalCase{ID: fmt.Sprintf("s%d-L%d-%d", rule.Seed, k, n), Start: a, Target: b}, k})
			n++
		}
	}
	return out
}

// BankDigestAlgo names the algorithm of BankDigest; it is part of every report.
const BankDigestAlgo = "sha256 over lines level\\0start\\0target\\n in case order; start/target = RenderPrefix; ids excluded"

// BankDigest hashes the ordered (level, start, target) content, not the ids.
func BankDigest(cs []BankCase) string {
	h := sha256.New()
	for _, c := range cs {
		fmt.Fprintf(h, "%d\x00%s\x00%s\n", c.Level, RenderPrefix(c.Start), RenderPrefix(c.Target))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// OverlapStats compares banks drawn for different seeds.
type OverlapStats struct {
	Cases, DistinctStart, DistinctPair, SharedPairsAcrossSeeds, InFixedBanks int
}

func BankOverlap(banks [][]BankCase) OverlapStats {
	var s OverlapStats
	starts, pairs := map[string]bool{}, map[string]int{}
	fixed := fixedBankKeys()
	for _, b := range banks {
		for _, c := range b {
			s.Cases++
			starts[RenderPrefix(c.Start)] = true
			pk := pairKey(c.Start, c.Target)
			pairs[pk]++
			if fixed[pk] {
				s.InFixedBanks++
			}
		}
	}
	s.DistinctStart, s.DistinctPair = len(starts), len(pairs)
	keys := make([]string, 0, len(pairs))
	for k := range pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if pairs[k] > 1 {
			s.SharedPairsAcrossSeeds++
		}
	}
	return s
}

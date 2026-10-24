package main

// G1 (new protocol, mathzero-next). Bank whose Start and Target (by Key, the
// structural associative-commutative normalization in mathring.go; NOT algebraic
// equivalence) are disjoint from a PRIOR set and unique inside the bank. Rule and
// seed are frozen in pilots/g1/DESIGN-G1-frozen.md.

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

const G1Seed int64 = 20261024

// G1Rule: MaxAttemptsPerLevel 0 means the declared formula 200*PerLevel+1000.
type G1Rule struct {
	Seed                int64
	PerLevel            int
	Levels              []int
	MaxAttemptsPerLevel int
}

type G1LevelStats struct {
	Level, Attempts, Accepted                       int
	EqualKey, StartPrior, TargetPrior, RepeatInBank int
}

type G1GenStats struct {
	PerLevel  []G1LevelStats
	PriorKeys int
}

// ParsePrefix inverts RenderPrefix. It returns an error for anything else.
func ParsePrefix(s string) (*Expr, error) {
	e, rest, err := g1parsePrefix(s)
	if err != nil {
		return nil, err
	}
	if rest != "" {
		return nil, fmt.Errorf("trailing %q", rest)
	}
	return e, nil
}

func g1parsePrefix(s string) (*Expr, string, error) {
	switch {
	case strings.HasPrefix(s, "neg("):
		a, rest, err := g1parsePrefix(s[4:])
		if err != nil || !strings.HasPrefix(rest, ")") {
			return nil, "", fmt.Errorf("bad neg in %q", s)
		}
		return ENeg(a), rest[1:], nil
	case strings.HasPrefix(s, "add("), strings.HasPrefix(s, "mul("):
		op := s[:3]
		a, rest, err := g1parsePrefix(s[4:])
		if err != nil || !strings.HasPrefix(rest, ",") {
			return nil, "", fmt.Errorf("bad %s in %q", op, s)
		}
		b, rest2, err := g1parsePrefix(rest[1:])
		if err != nil || !strings.HasPrefix(rest2, ")") {
			return nil, "", fmt.Errorf("bad %s in %q", op, s)
		}
		if op == "add" {
			return EAdd(a, b), rest2[1:], nil
		}
		return EMul(a, b), rest2[1:], nil
	case strings.HasPrefix(s, "c"):
		i := 1
		if i < len(s) && s[i] == '-' {
			i++
		}
		j := i
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		if j == i {
			return nil, "", fmt.Errorf("bad const in %q", s)
		}
		n, err := strconv.Atoi(s[1:j])
		if err != nil {
			return nil, "", err
		}
		return C(n), s[j:], nil
	case strings.HasPrefix(s, "v") && len(s) >= 2:
		return V(s[1:2]), s[2:], nil
	}
	return nil, "", fmt.Errorf("cannot parse %q", s)
}

// hindsightExcludedByKey: true when the Key of the hindsight Start or of its real
// substitute target is in keys, or when either side cannot be parsed (fail closed).
func hindsightExcludedByKey(h *HindsightPair, keys map[string]bool) bool {
	a, errA := ParsePrefix(h.Start)
	b, errB := ParsePrefix(h.Target)
	if errA != nil || errB != nil {
		return true
	}
	return keys[Key(a)] || keys[Key(b)]
}

// G1PriorKeys: Key of Start and Target of every problem in levels.json and
// blind.json (levels 2..8) and of the P3 secondary bank (GenBank seed 9001,
// levels 6,7,8 x 30). The P3 primary bank is the L6-8 part of levels.json.
func G1PriorKeys() map[string]bool {
	m := map[string]bool{}
	for _, f := range []string{"levels.json", "blind.json"} {
		for k := 2; k <= 8; k++ {
			for _, p := range LoadProblems(f, fmt.Sprint(k)) {
				m[Key(p.Cur)], m[Key(p.Target)] = true, true
			}
		}
	}
	for _, c := range GenBank(BankRule{Seed: 9001, PerLevel: 30, Levels: []int{6, 7, 8}}) {
		m[Key(c.Start)], m[Key(c.Target)] = true, true
	}
	return m
}

// GenBankG1 is deterministic in (rule, prior). If a level cannot reach PerLevel
// within its attempt limit it FAILS CLOSED: the error is returned together with the
// partial cases and the statistics; nothing is relaxed.
func GenBankG1(rule G1Rule, prior map[string]bool) ([]BankCase, G1GenStats, error) {
	r := rand.New(rand.NewSource(rule.Seed))
	seen := map[string]bool{}
	st := G1GenStats{PriorKeys: len(prior)}
	var out []BankCase
	for _, k := range rule.Levels {
		limit := rule.MaxAttemptsPerLevel
		if limit == 0 {
			limit = 200*rule.PerLevel + 1000
		}
		ls := G1LevelStats{Level: k}
		for ls.Attempts < limit && ls.Accepted < rule.PerLevel {
			ls.Attempts++
			a, b := genTrain(r, k)
			ka, kb := Key(a), Key(b)
			switch {
			case ka == kb:
				ls.EqualKey++
			case prior[ka]:
				ls.StartPrior++
			case prior[kb]:
				ls.TargetPrior++
			case seen[ka] || seen[kb]:
				ls.RepeatInBank++
			default:
				seen[ka], seen[kb] = true, true
				out = append(out, BankCase{EvalCase{ID: fmt.Sprintf("g1-%d-L%d-%d", rule.Seed, k, ls.Accepted), Start: a, Target: b}, k})
				ls.Accepted++
			}
		}
		st.PerLevel = append(st.PerLevel, ls)
		if ls.Accepted < rule.PerLevel {
			return out, st, fmt.Errorf("G1 bank infeasible at level %d: %d/%d accepted in %d attempts (failing closed, rule not relaxed)", k, ls.Accepted, rule.PerLevel, ls.Attempts)
		}
	}
	return out, st, nil
}

// G1KeySetOf: Key of every Start and Target of the bank, as one set.
func G1KeySetOf(cs []BankCase) map[string]bool {
	m := map[string]bool{}
	for _, c := range cs {
		m[Key(c.Start)], m[Key(c.Target)] = true, true
	}
	return m
}

// ---- A1 bank G2 (frozen in pilots/a1/DESIGN-A1-frozen.md) ----

const G2Seed int64 = 20261025

// G2PriorKeys = G1PriorKeys + Key of Start and Target of the frozen G1 bank. It
// fails closed if the regenerated G1 bank does not match its frozen digest.
func G2PriorKeys() (map[string]bool, error) {
	prior := G1PriorKeys()
	g1, _, err := GenBankG1(G1Rule{Seed: G1Seed, PerLevel: 30, Levels: []int{6, 7, 8}}, prior)
	if err != nil {
		return nil, err
	}
	if BankDigest(g1) != G1FrozenDigest {
		return nil, fmt.Errorf("G1 bank digest %s differs from frozen %s", BankDigest(g1), G1FrozenDigest)
	}
	for k := range G1KeySetOf(g1) {
		prior[k] = true
	}
	return prior, nil
}

package main

// A1 (new protocol, mathzero-next). Training with a FIXED level schedule, a
// problem RNG that never depends on the search, a per-(seed,round) SGD RNG, and
// two arms: A (the learner also guides the search) and F (the guide is a frozen
// deep copy of the initial network). TrainRounds (P3/G1) is untouched.
// Design: pilots/a1/DESIGN-A1-frozen.md.

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

type A1Config struct {
	Seed        int64
	Schedule    []int // level per round, fixed in advance
	PerRound    int
	Cap         int
	FrozenGuide bool            // false: arm A; true: arm F
	ExcludeKeys map[string]bool // Key of every Start and Target of the evaluation bank
}

type A1Result struct {
	Log                       TrainLog
	GuideSHAPre, GuideSHAPost string // arm F only
}

// CloneMLP is a deep copy by JSON round trip (weights, Level, Round; no optimizer
// state: a guide never trains).
func CloneMLP(n *MLP) (*MLP, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	c := &MLP{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return c, nil
}

// a1SGDSeed: one SGD stream per (seed, round), the same for both arms.
func a1SGDSeed(seed int64, round int) int64 { return seed*1000003 + int64(round)*7919 + 17 }

func TrainRoundsA1(net *MLP, cfg A1Config) (A1Result, error) {
	var res A1Result
	if len(cfg.Schedule) < 1 || cfg.PerRound < 1 || cfg.Cap < 2 || cfg.ExcludeKeys == nil {
		return res, fmt.Errorf("invalid A1 config (schedule, perRound, cap, ExcludeKeys required)")
	}
	guide := net
	if cfg.FrozenGuide {
		g, err := CloneMLP(net)
		if err != nil {
			return res, err
		}
		guide = g
		res.GuideSHAPre = WeightsSHA(guide)
	}
	prevT, prevL := SearchTraining, LemmaLib
	SearchTraining, LemmaLib = true, nil
	defer func() { SearchTraining, LemmaLib = prevT, prevL }()
	rp := rand.New(rand.NewSource(cfg.Seed)) // problems only
	for ri, level := range cfg.Schedule {
		round := ri + 1
		var buf []Sample
		rl := TrainRoundLog{Round: round, Level: level, Problems: cfg.PerRound}
		for i := 0; i < cfg.PerRound; i++ {
			var a, b *Expr
			ok := false
			for attempt := 0; attempt < 1000; attempt++ {
				a, b = genTrain(rp, level)
				rl.Attempts++
				if cfg.ExcludeKeys[Key(a)] || cfg.ExcludeKeys[Key(b)] {
					rl.KeyRejected++
					continue
				}
				ok = true
				break
			}
			if !ok {
				return res, fmt.Errorf("no training problem outside the evaluation key set at level %d (round %d)", level, round)
			}
			res.Log.Pairs = append(res.Log.Pairs, pairKey(a, b))
			res.Log.Starts = append(res.Log.Starts, Key(a))
			res.Log.Targets = append(res.Log.Targets, Key(b))
			res.Log.RenderedProblems = append(res.Log.RenderedProblems, RenderPrefix(a)+"|"+RenderPrefix(b))
			r := guidedSolveBidirectional(guide, a, b, cfg.Cap, true)
			if r.Hindsight != nil {
				if hindsightExcludedByKey(r.Hindsight, cfg.ExcludeKeys) {
					rl.HindsightDropped++
				} else {
					res.Log.HindsightKeptRendered = append(res.Log.HindsightKeptRendered, r.Hindsight.Start+"|"+r.Hindsight.Target)
					rl.HindsightKept++
					buf = append(buf, r.Samples...)
				}
			} else if r.OK {
				rl.Solved++
				buf = append(buf, r.Samples...)
			}
			rl.Samples = len(buf)
		}
		rs := rand.New(rand.NewSource(a1SGDSeed(cfg.Seed, round)))
		for ep := 0; ep < 3; ep++ {
			rs.Shuffle(len(buf), func(i, j int) { buf[i], buf[j] = buf[j], buf[i] })
			for _, s := range buf {
				net.TrainCE(s.Feats, s.Target, 0.003)
			}
		}
		net.Round++
		res.Log.Rounds = append(res.Log.Rounds, rl)
	}
	if cfg.FrozenGuide {
		res.GuideSHAPost = WeightsSHA(guide)
	}
	return res, nil
}

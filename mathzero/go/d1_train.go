package main

// P3 (new code, mathzero-next). Self-play + imitation training with explicit
// configuration, no library, no files, no checkpoint selection. Logic follows
// Cycle's training step (bidirectional guided search, TrainCE lr 0.003, 3
// shuffled epochs, level rises when >=80% of the round is solved) but is NEW
// code; it does not evaluate and never reads an evaluation bank except to
// EXCLUDE its pairs from training.

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
)

type TrainConfig struct {
	Seed     int64
	Rounds   int
	PerRound int
	Cap      int
	Exclude  map[string]bool // pairKey of evaluation pairs (AC-normalized, either orientation)
	// ExcludeRender holds RenderPrefix(start)+"|"+RenderPrefix(target) of evaluation
	// pairs (either orientation); hindsight pairs found in it are dropped.
	ExcludeRender map[string]bool
	// G1 (new protocol). When ExcludeKeys is non-nil, any training problem whose
	// Start or Target has a Key in ExcludeKeys is rejected and regenerated, a
	// hindsight pair is dropped when the Key of its Start or of its REAL
	// substitute target is in ExcludeKeys (or the pair cannot be parsed), and the
	// rendered problems and kept hindsight pairs are recorded in the log. nil =
	// P3 behaviour, bit for bit.
	ExcludeKeys map[string]bool
}

type TrainRoundLog struct {
	Round, Level, Solved, Problems, Samples, HindsightKept, HindsightDropped int
	// G1 only (omitted otherwise): attempts drawn and rejected by the Key rule.
	Attempts, KeyRejected int `json:",omitempty"`
}

type TrainLog struct {
	Rounds  []TrainRoundLog
	Pairs   []string // pairKey of every training problem, in order
	Starts  []string // Key of each training start
	Targets []string
	// G1 only (omitted otherwise): rendered training problems ("start|target") and
	// kept hindsight pairs, for verification with a normalization independent of Key.
	RenderedProblems      []string `json:",omitempty"`
	HindsightKeptRendered []string `json:",omitempty"`
}

// WeightsSHA hashes the exported weights, Level and Round (not optimizer state).
func WeightsSHA(n *MLP) string {
	b, _ := json.Marshal(n)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func reversedKey(a, b *Expr) string { return pairKey(b, a) }

// ExcludeSetFor builds the exclusion set from evaluation cases (both orientations).
func ExcludeSetFor(cs []EvalCase) map[string]bool {
	m := map[string]bool{}
	for _, c := range cs {
		m[pairKey(c.Start, c.Target)] = true
		m[reversedKey(c.Start, c.Target)] = true
	}
	return m
}

// RenderExcludeFor builds ExcludeRender from evaluation cases.
func RenderExcludeFor(cs []EvalCase) map[string]bool {
	m := map[string]bool{}
	for _, c := range cs {
		a, b := RenderPrefix(c.Start), RenderPrefix(c.Target)
		m[a+"|"+b], m[b+"|"+a] = true, true
	}
	return m
}

// TrainRounds mutates net. LemmaLib is forced to nil during training and
// SearchTraining is set; both are restored on return.
func TrainRounds(net *MLP, cfg TrainConfig) (TrainLog, error) {
	var log TrainLog
	if cfg.Rounds < 1 || cfg.PerRound < 1 || cfg.Cap < 2 {
		return log, fmt.Errorf("invalid rounds/perRound/cap")
	}
	prevT, prevL := SearchTraining, LemmaLib
	SearchTraining, LemmaLib = true, nil
	defer func() { SearchTraining, LemmaLib = prevT, prevL }()
	r := rand.New(rand.NewSource(cfg.Seed))
	for round := 1; round <= cfg.Rounds; round++ {
		var buf []Sample
		rl := TrainRoundLog{Round: round, Level: net.Level, Problems: cfg.PerRound}
		for i := 0; i < cfg.PerRound; i++ {
			var a, b *Expr
			ok := false
			for attempt := 0; attempt < 1000; attempt++ {
				a, b = genTrain(r, net.Level)
				if cfg.ExcludeKeys != nil {
					rl.Attempts++
				}
				if cfg.Exclude[pairKey(a, b)] {
					continue
				}
				if cfg.ExcludeKeys != nil && (cfg.ExcludeKeys[Key(a)] || cfg.ExcludeKeys[Key(b)]) {
					rl.KeyRejected++
					continue
				}
				ok = true
				break
			}
			if !ok {
				return log, fmt.Errorf("no training pair outside the evaluation set at level %d", net.Level)
			}
			log.Pairs = append(log.Pairs, pairKey(a, b))
			log.Starts = append(log.Starts, Key(a))
			log.Targets = append(log.Targets, Key(b))
			if cfg.ExcludeKeys != nil {
				log.RenderedProblems = append(log.RenderedProblems, RenderPrefix(a)+"|"+RenderPrefix(b))
			}
			res := guidedSolveBidirectional(net, a, b, cfg.Cap, true)
			if res.Hindsight != nil {
				if cfg.ExcludeRender[res.Hindsight.Start+"|"+res.Hindsight.Target] || (cfg.ExcludeKeys != nil && hindsightExcludedByKey(res.Hindsight, cfg.ExcludeKeys)) {
					rl.HindsightDropped++
				} else {
					if cfg.ExcludeKeys != nil {
						log.HindsightKeptRendered = append(log.HindsightKeptRendered, res.Hindsight.Start+"|"+res.Hindsight.Target)
					}
					rl.HindsightKept++
					buf = append(buf, res.Samples...)
				}
			} else if res.OK {
				rl.Solved++
				buf = append(buf, res.Samples...)
			}
			rl.Samples = len(buf)
		}
		for ep := 0; ep < 3; ep++ {
			r.Shuffle(len(buf), func(i, j int) { buf[i], buf[j] = buf[j], buf[i] })
			for _, s := range buf {
				net.TrainCE(s.Feats, s.Target, 0.003)
			}
		}
		net.Round++
		if float64(rl.Solved)/float64(cfg.PerRound) >= 0.8 && net.Level < 8 {
			net.Level++
		}
		log.Rounds = append(log.Rounds, rl)
	}
	return log, nil
}

// TrainEvalOverlap counts training problems whose normalized Start / Target /
// pair appears in the evaluation cases (normalization: Key, AC-normalized).
type TrainEvalOverlap struct {
	TrainProblems, StartInEval, TargetInEval, PairInEval int
}

func OverlapTrainEval(log TrainLog, eval []EvalCase) TrainEvalOverlap {
	es, et, ep := map[string]bool{}, map[string]bool{}, ExcludeSetFor(eval)
	for _, c := range eval {
		es[Key(c.Start)], et[Key(c.Target)] = true, true
	}
	o := TrainEvalOverlap{TrainProblems: len(log.Pairs)}
	for i := range log.Pairs {
		if es[log.Starts[i]] {
			o.StartInEval++
		}
		if et[log.Targets[i]] {
			o.TargetInEval++
		}
		if ep[log.Pairs[i]] {
			o.PairInEval++
		}
	}
	return o
}

//go:build !js

package main

// G3: bank for the A1C confirmation (seeds 19..26), same generator and levels as
// G2 (L6,7,8 x 30). Its seed (20261027) was fixed before generation.
// PRIOR = G2PriorKeys + Key of Start and Target of the frozen G2 bank. The bank
// never depends on any model or result. Training problems are kept out of the bank
// by A1's own mechanism (the bank is frozen first and TrainRoundsA1 rejects any
// problem whose Start or Target Key is in it, and drops hindsight pairs likewise).
// Disjointness is by Key (structural AC), not algebraic equivalence.

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const G3Seed int64 = 20261027

// G3FrozenDigest: CANDIDATE (generated once, deterministic, no model); frozen only after independent review.
const G3FrozenDigest = "9df061e0f89aae8736ffda6e8b562d90dede192d0d9e2f928440a3f50bab7a19"

// a1BankTag names the bank of a rule: "g3" for the G3 seed, otherwise "g2"
// (default, unchanged behaviour including tests with custom rules).
func a1BankTag(r G1Rule) string {
	if r.Seed == G3Seed {
		return "g3"
	}
	return "g2"
}

func G3PriorKeys() (map[string]bool, error) {
	prior, err := G2PriorKeys()
	if err != nil {
		return nil, err
	}
	g2, _, err := GenBankG1(G1Rule{Seed: G2Seed, PerLevel: 30, Levels: []int{6, 7, 8}}, prior)
	if err != nil {
		return nil, err
	}
	if BankDigest(g2) != G2FrozenDigest {
		return nil, fmt.Errorf("G2 bank digest %s differs from frozen %s", BankDigest(g2), G2FrozenDigest)
	}
	out := make(map[string]bool, len(prior)+180)
	for k := range prior {
		out[k] = true
	}
	for k := range G1KeySetOf(g2) {
		out[k] = true
	}
	return out, nil
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1g3bank" {
		if err := runD1G3Bank(); err != nil {
			fmt.Fprintln(os.Stderr, "d1g3bank:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

// runD1G3Bank: generation only (no model, no evaluation). Fails closed.
func runD1G3Bank() error {
	t0 := time.Now()
	prior, err := G3PriorKeys()
	if err != nil {
		return err
	}
	bank, st, err := GenBankG1(G1Rule{Seed: G3Seed, PerLevel: 30, Levels: []int{6, 7, 8}}, prior)
	out := map[string]any{"seed": G3Seed, "prior_keys": st.PriorKeys, "generation_stats": st.PerLevel, "cases": len(bank), "digest": BankDigest(bank), "seconds": time.Since(t0).Seconds()}
	var cs []string
	for _, c := range bank {
		cs = append(cs, fmt.Sprintf("%d|%s|%s", c.Level, RenderPrefix(c.Start), RenderPrefix(c.Target)))
	}
	out["rendered"] = cs
	if err != nil {
		out["error"] = err.Error()
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", " ")
	e.Encode(out)
	return err
}

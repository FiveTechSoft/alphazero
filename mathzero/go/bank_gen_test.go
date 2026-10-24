package main

import "testing"

func TestGenBankDeterministicAndDeclared(t *testing.T) {
	rule := BankRule{Seed: 5, PerLevel: 6, Levels: []int{4, 6}}
	a, b := GenBank(rule), GenBank(rule)
	if len(a) != 12 || BankDigest(a) != BankDigest(b) {
		t.Fatalf("len %d or digest differs", len(a))
	}
	if BankDigest(a) == BankDigest(GenBank(BankRule{Seed: 6, PerLevel: 6, Levels: []int{4, 6}})) {
		t.Fatal("different seeds gave the same bank")
	}
	per := map[int]int{}
	for _, c := range a {
		per[c.Level]++
	}
	if per[4] != 6 || per[6] != 6 {
		t.Fatalf("composition %v", per)
	}
}

func TestGenBankExcludesFixedAndDuplicates(t *testing.T) {
	cs := GenBank(BankRule{Seed: 9, PerLevel: 20, Levels: []int{2, 4}})
	fixed := fixedBankKeys()
	seen := map[string]bool{}
	for _, c := range cs {
		pk := pairKey(c.Start, c.Target)
		if fixed[pk] || seen[pk] {
			t.Fatalf("case %s is fixed or duplicated", c.ID)
		}
		seen[pk] = true
		if RenderPrefix(c.Start) == RenderPrefix(c.Target) {
			t.Fatal("identity case")
		}
	}
	if len(cs) != 40 {
		t.Fatalf("got %d", len(cs))
	}
}

func TestBankDigestDependsOnContentNotIDs(t *testing.T) {
	cs := GenBank(BankRule{Seed: 3, PerLevel: 4, Levels: []int{4}})
	d := BankDigest(cs)
	for i := range cs {
		cs[i].ID = "renamed"
	}
	if BankDigest(cs) != d {
		t.Fatal("digest depends on IDs")
	}
	if BankDigest(cs[:3]) == d {
		t.Fatal("digest ignores content")
	}
}

func TestBankOverlapCounts(t *testing.T) {
	a := GenBank(BankRule{Seed: 1, PerLevel: 8, Levels: []int{4}})
	s := BankOverlap([][]BankCase{a, a})
	if s.Cases != 16 || s.DistinctPair != 8 || s.SharedPairsAcrossSeeds != 8 || s.InFixedBanks != 0 {
		t.Fatalf("%+v", s)
	}
}

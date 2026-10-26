# dlg_selftrain — status
Experiment in progress (2026-10-26). Design: DESIGN.md.

Measured so far (seed 7001; dev = monitoring sets, test = reserved, evaluated
once per arm at its closeout):
- base adequacy gate: PASSED (dev_seen explicit accuracy 0.5574, n=183,
  threshold 0.30).
- b0 (base only, no self-training rounds): test_seen 0.715, test_novel 0.585,
  test_novel_long 0.537. These are the reference numbers the self arm must
  beat; they are NOT evidence of self-training improvement.
Details: results/base_b0_seed7001.json.
- sup (free-supervision reference, 3 rounds gold answers, seed 7001):
  test_seen 0.772 (+0.057 vs b0), test_novel 0.575 (-0.010),
  test_novel_long 0.527 (-0.010): more supervised rounds improved SEEN
  accuracy but not novel-compositional accuracy on this seed.
  Details: results/sup_seed7001.json.
- self (verified own answers, 3 rounds, seed 7001): test_seen 0.800,
  test_novel 0.608, test_novel_long 0.523. Point differences vs b0 on this
  seed: +0.085 seen, +0.023 novel, -0.014 novel_long; vs sup: +0.028 seen,
  +0.033 novel, -0.004 novel_long. One seed, no uncertainty estimate yet;
  H1 (self > b0 on test_novel on BOTH seeds) is open until seed 7002.
  Details: results/self_seed7001.json.
- selfuf (control: own samples WITHOUT verification, 3 rounds, seed 7001):
  test_seen 0.725, test_novel 0.593, test_novel_long 0.540; unfiltered data
  was ~45-55% wrong answers (first_sample_acc 0.546/0.455/0.545).
  Details: results/selfuf_seed7001.json.
Seed 7001 complete. Exact TEST counts (single seed, no uncertainty estimate):

  arm    | seen total (pron/expl) | novel total (pron/expl) | long total (pron/expl)
  b0     | 286/400 (144/150, 142/250) | 234/400 (158/170, 76/230) | 161/300 (96/112, 65/188)
  sup    | 309/400 (143/150, 166/250) | 230/400 (168/170, 62/230) | 158/300 (110/112, 48/188)
  self   | 320/400 (147/150, 173/250) | 243/400 (168/170, 75/230) | 157/300 (108/112, 49/188)
  selfuf | 290/400 (140/150, 150/250) | 237/400 (163/170, 74/230) | 162/300 (103/112, 59/188)

Decomposition that the aggregates hide: on test_novel every arm answers almost
all pronoun questions correctly (158-168/170); self's net +9 over b0 comes
from pronouns (+10) while EXPLICIT novel questions do not improve in any arm
vs b0 (76 -> sup 62, self 75, selfuf 74 out of 230). On long dialogues the
pattern inverts for self: better pronouns (+12) but worse explicit (-16).
Hypotheses, reported per split (no retrospective re-specification):
- H1 (self > b0 on test_novel, both seeds): seed 7001 in favor on the AGGREGATE
  (+0.023), driven by pronoun questions, not by the explicit compositional
  questions the split was built to measure. Seed 7002 pending.
- H2 (self <= sup): against on this seed (self +0.033 novel aggregate over sup).
- H3 (self > selfuf): split-dependent on this seed: novel aggregate +0.015 in
  favor, novel_long -0.017 against. No stable verifier advantage is claimed.
Open honest reading (n=1): rounds with or without verification did not improve
explicit compositional generalization; self improved pronoun reference
tracking in new contexts. Seed 7002 (partial; base gate passed, dev_seen expl 0.5464 n=183):
  arm | seen total (pron/expl) | novel total (pron/expl) | long total (pron/expl)
  b0  | 267/400 (133/150, 134/250) | 228/400 (152/170, 76/230) | 158/300 (94/112, 64/188)
  sup | 297/400 (147/150, 150/250) | 219/400 (158/170, 61/230) | 144/300 (106/112, 38/188)
  Same pattern as seed 7001 for sup: better seen, WORSE explicit novel
  (76->61/230) and explicit long (64->38/188). self/selfuf 7002 pending.
  Details: results/base_b0_sup_seed7002.json.

## Seed 7002 - arm `self` (TEST, single evaluation)

| split | total | pronouns | explicit |
|---|---|---|---|
| test_seen | 292/400 | 150/150 | 142/250 |
| test_novel | 211/400 | 167/170 | 44/230 |
| test_novel_long | 153/300 | 112/112 | 41/188 |

Acceptance (verifier pass rate per round): 0.7225 / 0.775 / 0.770; kept texts 1025 / 1127 / 1152.

Comparison vs base arm b0 on the same seed (test_novel 228/400 = 152/170 pron, 76/230 expl):
self is BELOW b0 on test_novel in seed 7002 (-17 answers). The decomposition shows the same
directional split as seed 7001 but stronger: pronouns improve (+15: 167 vs 152) while explicit
answers regress sharply (-32: 44 vs 76). H1 (self > b0 on test_novel in both seeds) is therefore
NOT supported: seed 7001 was +9, seed 7002 is -17.

H2 (self <= sup) had NO prespecified split, so we report it as a descriptive per-split
comparison, not a confirmatory test. Direction differs by split and seed:
- seed 7001: test_seen self 320 vs sup 309 (self higher); test_novel 243 vs 230 (self higher);
  test_novel_long 157 vs 158 (sup higher by 1).
- seed 7002: test_seen 292 vs 297 (sup higher); test_novel 211 vs 219 (sup higher);
  test_novel_long 153 vs 144 (self higher).
No consistent ordering between self and sup across splits or seeds.

Aggregates only; raw fail examples are not published.

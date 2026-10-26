# Final report: self-training (STaR-style) with a programmatic verifier on Spanish state-tracking dialogues

Experiment branch: exp/dialogue-selftrain. Model: char-level GPT trained FROM SCRATCH per seed
(841,600 params, D128 L4 H4 CTX352). It is NOT an evolution of ES4's weights; ES4 was untouched.
Two seeds (7001, 7002); one reserved TEST evaluation per arm; dev never used for selection or tuning.
This is an exploratory 2-seed study. No RSI claims are made.

## Status: INCOMPLETE RUN

Arms b0 (base only), sup (supervised control, equal gradient steps) and self (verified self-training)
completed on both seeds. Arm selfuf (unfiltered self-training control) completed on seed 7001 only;
on seed 7002 it was interrupted twice by process kills (see Interruptions) and the run was closed by
decision without consuming its reserved TEST. The self vs selfuf comparison on seed 7002 is pending;
no results are imputed for the missing arm.

## TEST results (single reserved evaluation per arm; exact counts)

Denominators: test_seen 400 (150 pronoun / 250 explicit); test_novel 400 (170/230);
test_novel_long 300 (112/188).

### Seed 7001

| arm | seen | novel | long |
|---|---|---|---|
| b0 | 286/400 (144/150, 142/250) | 234/400 (158/170, 76/230) | 161/300 (96/112, 65/188) |
| sup | 309/400 (143/150, 166/250) | 230/400 (168/170, 62/230) | 158/300 (110/112, 48/188) |
| self | 320/400 (147/150, 173/250) | 243/400 (168/170, 75/230) | 157/300 (108/112, 49/188) |
| selfuf | 290/400 (140/150, 150/250) | 237/400 (163/170, 74/230) | 162/300 (103/112, 59/188) |

### Seed 7002

| arm | seen | novel | long |
|---|---|---|---|
| b0 | 267/400 (133/150, 134/250) | 228/400 (152/170, 76/230) | 158/300 (94/112, 64/188) |
| sup | 297/400 (147/150, 150/250) | 219/400 (158/170, 61/230) | 144/300 (106/112, 38/188) |
| self | 292/400 (150/150, 142/250) | 211/400 (167/170, 44/230) | 153/300 (112/112, 41/188) |
| selfuf | INCOMPLETE (no TEST consumed) | - | - |

## Hypotheses

- H1 (self > b0 on test_novel in both seeds): NOT SUPPORTED. Seed 7001: 243 vs 234 (+9).
  Seed 7002: 211 vs 228 (-17). Per-type decomposition of the aggregate differences:
  seed 7001 pronouns +10 / explicit -1; seed 7002 pronouns +15 / explicit -32.
  Pronoun-type novel answers improved under self in both seeds; explicit novel answers did not
  improve in ANY arm vs b0 in either seed.
- H2 (self <= sup): NO split was prespecified for H2, so this is a descriptive per-split
  comparison, not a confirmatory test. Directions: seed 7001 seen 320>309, novel 243>230,
  long 157<158; seed 7002 seen 292<297, novel 211<219, long 153>144.
  No consistent ordering between self and sup across splits or seeds.
- H3 (self > selfuf; value of the verifier): evaluable on seed 7001 only, descriptively:
  seen +30 (320 vs 290), novel +6 (243 vs 237), long -5 (157 vs 162).
  NOT REPLICATED: the seed-7002 comparison is missing (arm incomplete). H3 remains unresolved.

## Acceptance stats (arm self, verifier pass rate per round)

Seed 7001: see results/self_seed7001.json. Seed 7002: 0.7225 / 0.775 / 0.770; kept 1025/1127/1152.
Runtime label verification used the independent parser (gen.check): 0 mismatches in all rounds.

## Interruptions and compute (measured, separated)

Three mid-run process kills occurred. Cause NOT DETERMINED in all three cases (sandbox restarts
happened in the same windows; no causal proof).

Kill 1 (base, SEED 7001 attempt 1): started 07:00:39 under v1 code, found dead at 07:16 with
last log "step 700/1500 loss 0.848". NO checkpoint existed (v1 saved only at completion), so
recovery was a RESTART from scratch, not a resume: attempt 2 launched 07:16:25 with the identical
preregistered config and completed 1500 steps at 07:31:06 in 876.6 s (metrics row has no from_step).
Seed 7002's base was a separate fresh start later (base_train from_step 0 at 08:42:36, 757.8 s) and
was never killed. Periodic atomic checkpointing (weights+Adam+RNG+step) was added after kill 1 and
its resume parity verified bit-exact; that mechanism is what the later selfuf resumes used.
Sources: contemporaneous OPLOG entry and metrics.jsonl base rows for both seeds.

Kill 2 (selfuf 7002 round0): dead at step 200/500. Resumed from preserved checkpoint step 200.
Kill 3 (selfuf 7002 round1): dead at step 100/500. No further retry; run closed by decision. Recovery evidence for the selfuf-7002 round-0
checkpoint (hashes, RNG state at step 200, identical regenerated corpus, continuous loss curve
0.240 -> 0.227 over steps 200-500) is consistent with a correct resume; it is NOT a bit-exact
end-to-end equivalence proof (weights/Adam not independently inspected).

Measured compute (seconds, from metrics logs):
- Discarded work: base seed-7001 attempt 1 ~372 (700/1500 steps killed, no checkpoint); selfuf 7002 round0 steps 0-200
  (pre-kill) ~103; selfuf 7002 round1 steps 0-100 (pre-kill) ~50.
- Completed: base 7001: 876.6; base 7002: 757.8; b0 evals 46.7 / 44.3; sup 920.0 / 881.1;
  self 1146.6 / 1152.2; selfuf 7001: 939.0; selfuf 7002 partial: round0 resumed 200-500 in 154.4
  plus generation/eval overheads. Total wall-clock for completed arms ~5.9k s of training.

## Limits

- 2 seeds, exploratory; per-type cells are small (112-250 items); differences of a few counts
  are described without noise or causality claims.
- Residual dev/test collision: 3 distinct prompts: 2 in dev_novel ∩ test_novel and 1
  DIFFERENT prompt in dev_novel ∩ test_novel_long (none shared between the two collisions;
  verified by direct recomputation from the frozen generator seeds 91002/92002/92003);
  declared as a limit; split sets frozen (SPLIT_FIXTURE_V2, D0 sha
  7c55384dfe9dab31421a1c6bedba0a4be3d8dccf15de5dbbe6315914b947114d).
- sup uses gold labels and is the reference of THIS training setup, not a theoretical ceiling.
- All training data paths (round prompts, sup gold, self-verified corpus) exclude held-out pairs
  (verified 0/800 per round, v2 fixture).
- Raw metrics with fail examples are not published; aggregates only.

## What remains open

selfuf seed 7002 from its preserved checkpoints (round1 step 100, evidence/selfuf7002_kill_r1/,
plus round-0 kill-point evidence in evidence/selfuf7002_kill_r0/), its single reserved TEST,
and with it the H3 replication. Checkpoints, logs and code are preserved for that future run.

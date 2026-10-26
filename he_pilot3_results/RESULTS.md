# HypothesisEngine pilot 3 - TEST results (seeds 601-648) and accounting
Frozen protocol/code: he_pilot3/ at the public pre-TEST snapshot (PREREG.md, main.go, analyze3.py, run_pilot3.sh, HASHES.txt). Domain: synthetic program search; selector is a hand-written rule or random; not neural, not RSI, not general learning.

## Preregistered result (analysis3.txt; mean audit solved of 300, 48 paired seeds, 3 arms, each arm-seed hard-capped at exactly 6,000,000 decision expansions)
Arm means (AuditSolve): excl_random 161.83 | freqrule 171.98 | random 160.27.
PRIMARY (only one) freqrule - excl_random: +10.15, 95% bootstrap CI [3.75, 16.56]. Rule: lower bound > 0, so the criterion is met. In this synthetic benchmark, ordering the non-deprioritised candidates by corpus frequency is associated with more audit-solved tasks than ordering them randomly, including trajectory feedback (the corpus depends on earlier acceptances).
This is NOT neural learning, NOT RSI, NOT general learning: freqrule is a hand-written rule, no model was trained.

## Secondary (descriptive, no multiplicity adjustment, no causal reading)
excl_random - random: +1.56 [-3.62, 6.48]. freqrule - random: +11.71 [5.29, 18.12].
Per-arm diagnostics (table order is alphabetical: excl_random, freqrule, random; it is not the execution order, which within each seed was freqrule, excl_random, random, with no randomisation of the order; means over 48 seeds):
| arm | rounds | accepts | evals | final library size | fallback rounds | tie rounds | mean wall ms | mean ops |
|---|---|---|---|---|---|---|---|---|
| excl_random | 41.6 | 8.21 | 125.52 | 16.21 | 0.0 | 0.0 (not counted) | 1152.44 | 769442 |
| freqrule | 45.58 | 13.54 | 137.4 | 21.54 | 0.0 | 24.0 | 1185.12 | 968416 |
| random | 41.31 | 10.31 | 124.83 | 13.77 | 0.0 | 0.0 (not counted) | 1145.52 | 758581 |
Tie rounds are counted for freqrule only: on average 24 of its ~45.6 rounds per seed had a tie at the selection cut, resolved by random noise. So "ranking by frequency" means preferring candidates whose expansion appears in the corpus, with substantial random tie-breaking. Wall time is on one shared core and noisy.

## Limits
One synthetic world family, one budget, audit inside the same family. The design does not separate remove from redundancy as the source of the deprioritisation effect, and the excl_random - random contrast is descriptive only. The design was conditioned by pilot 2 results and an exploratory diagnostic of pilot 2 artifacts; a 3-seed smoke run (seeds 901-903, not TEST) was seen before the freeze (declaration: no design change was made after seeing it; not independently verifiable).

## Cost accounting
- TEST: 144 arm-seed runs x 6,000,000 = 864,000,000 decision expansions (Spent = 6,000,000 verified in every row by run_pilot3.sh and analyze3.py).
- Fixtures and smoke before the freeze (seeds 901-903, not TEST, not evidence): 198,000,000 expansions (33 complete runs of 6M) plus 12,000,000 expansions of additional fixture runs = 210,000,000.
- Not counted in those numbers: audit evaluation (300 tasks per seed), pool generation, features, compilation, the exploratory script, wall time. Mean wall time per arm-seed run was about 1.15-1.19 s; whole TEST about 3 minutes on one local core. No server, no paid API, no training.
- Interruptions: reported execution details (not independently verified): one continuous execution of run_pilot3.sh, no interruption and no retries. The four complete chunk logs are consistent with this; they do not prove the absence of retries.

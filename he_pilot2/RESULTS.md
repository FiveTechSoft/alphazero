# HypothesisEngine pilot 2 - TEST results (seeds 501-548) and accounting
Frozen protocol/code/weights: PREREG.md, main.go, weights.json (published before the test; see he_pilot2 pre-test commit). Domain: synthetic program search; selector is a ridge model; not neural, not RSI.

## Result (mean audit solved of 300, 48 paired seeds, 6 arms, each arm hard-capped at exactly 6,000,000 decision expansions per seed)
freqrule 172.10 | transfer 156.35 | ucb 155.54 | random 154.85 | learned (online) 154.67 | transfer_shuf 146.33
Primary contrast (preregistered, only one): transfer - random = +1.50, 95% bootstrap CI [-5.19, +7.88]. The criterion (lower bound > 0) is not met: no advantage of the offline-trained selector over random was demonstrated. The interval does not establish equivalence either; effects from about -5 to +8 solved tasks are compatible with the data.
Descriptive, no multiplicity adjustment, interpretation limited:
- transfer - transfer_shuf +10.02 [4.04, 16.31]; transfer - freqrule -15.75 [-22.42, -8.98]; transfer - learned +1.69 [-4.00, 7.62]; transfer - ucb +0.81 [-5.33, 6.52]; learned - random -0.19 [-6.62, 6.17]; freqrule - random +17.25 [10.56, 24.04].
- The transfer vs transfer_shuf difference does not by itself show that shuffled < random, nor isolate data as the cause. freqrule's higher mean (same sign as pilot 1, now at equal expansions) is an observation and a hypothesis for later work.

## Cost accounting
- Test budget: every arm spent exactly 6,000,000 expansions per seed (checked in the output); 48 seeds = 288,000,000 expansions per arm; 6 arms = 1,728,000,000 (1.728 billion) expansions in the TEST final chunks, excluding unmeasured partial attempts.
- Offline training of the transfer weights: 192M expansions (32 seeds x 6M), excluded from the test budget. 192M/6M = 32x one run's budget; relative to the whole test per arm (288M) it is 0.67x. Total compute is NOT equal between transfer arms and the others.
- Not in the expansion budget: ridge fit, pool generation, features, audit. Per-arm mean wall time ~0.96-1.03 s per seed and ridge/feature op counter 0.70-1.11M (online learned highest) are recorded in final.json/analysis; wall time was measured on a shared single core and is noisy.
- Interruptions: (1) a first launch command was blocked before running (no compute). (2) A second launch of all chunks was killed at the tool time limit: chunk 501-512 completed and was kept; part of chunk 513-524 ran and its output was overwritten when that chunk was rerun. (3) A background relaunch was terminated by the environment. Compute spent in killed/overwritten partial runs was not measured and is unknown; determinism does not make it zero. Final results come from four complete 12-seed chunk runs (501-512, 513-524, 525-536, 537-548); final.json/final.log are their concatenation. Some seeds were thus computed more than once; no setting was changed and no results were inspected before the merge (both are declarations, not independently verified). Logs of the interrupted/partial attempts were not kept (overwritten or lost); only the four complete chunk outputs exist.
- Training/DEV: collect seeds 201-232; DEV 401-408 run once as debugging, not evidence (declaration only).

## Limits
One synthetic world family, one budget, one ridge feature set; the audit is inside the same world family (no arbitrary transfer); wall time not controlled; multiple descriptive contrasts unadjusted.

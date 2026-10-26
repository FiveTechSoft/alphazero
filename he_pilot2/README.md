# HypothesisEngine pilot 2 (PRE-TEST snapshot): offline-trained edit selector vs controls, synthetic program-search domain
STATUS: code, protocol and weights frozen BEFORE the test run (seeds 501-548 not yet run at publication). Results will be added in a later commit.

Read first:
1. Cost: every arm gets exactly 6,000,000 decision expansions at test time (hard cap; a round cut by the cap is discarded, its spend counts). The transfer arms' weights were trained offline with 192M expansions (32x the test budget), EXCLUDED from that budget. Total compute is NOT equal between transfer and non-transfer arms. Ridge fit, pool generation, features and the audit are not in the expansion budget; wall time and a ridge/feature op counter are recorded per arm.
2. Question: does a prior trained on OTHER synthetic worlds (seeds 201-232) beat random selection at equal in-use budget on unseen worlds (seeds 501-548)? Not general learning, not neural, not RSI. The audit (300 tasks/seed) is within the same world family, not arbitrary transfer.
3. DEV seeds 401-408 (dev_v1.json/log) are debugging only, not evidence; run once, no changes afterwards (declaration only, not independently verified). Chronology of the freeze is by commit of this snapshot.
4. Pilot 1 (he_pilot/) was a negative result with unequal budgets; this pilot corrects the budget cap and pool determinism, it does not reuse pilot 1 data.

Files: main.go, go.mod (Go, no dependencies), PREREG.md, analyze.py, train.json (4848 labelled edits), weights.json, collect.log, dev_v1.json, dev_v1.log, HASHES.txt.
Reproduce offline fit: go build -o he . && ./he -mode collect -lo 201 -hi 232 -train train.json -weights weights.json (ridge lambda=1; shuffled control permutation seed 424242).
Test (after freeze): ./he -lo 501 -hi 548 -out final.json; python3 analyze.py final.json

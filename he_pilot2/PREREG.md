# HypothesisEngine pilot 2 - protocol (written before any test-seed run)
Question: does an edit selector whose weights were LEARNED OFFLINE on other worlds (frozen ridge, transfer) beat random
selection, at a strictly equal hard budget of decision expansions, on unseen worlds? Pilot 1 (online ridge from ~100-200 noisy
labels) showed no advantage; this tests the 'too little data per run' explanation. Synthetic domain; not neural; not RSI.
Changes vs pilot 1 (code otherwise the same): (a) HARD budget: evaluation is truncated at exactly 6,000,000 expansions; a round cut by the
budget is discarded (no acceptance, no label) but its spend counts, so every arm spends exactly 6,000,000 (verified in output);
(b) pool = deterministic function of (seed, round, library) so equal libraries get equal pools in all arms; (c) new arms transfer / transfer_shuf;
(d) exhaustive arm dropped; (e) wall time and a non-expansion op counter (ridge/feature inner loops) recorded per arm.
Training (offline, before test): mode collect, seeds 201-232, arm collect = evaluate all 24 candidates each round, accept a random
improving edit; labels = solved-gain on the 40-task batch; one ridge fit on all samples (weights.json; transfer_shuf = same fit on label-permuted data,
control for label information). Training cost (192M expansions, 32x the test budget) is NOT included in the test budget: total compute is NOT equal
between transfer arms and non-transfer arms; this is declared and is the price of the transfer question.
DEV (debug only, not evidence): seeds 401-408 once, no tuning after (dev_v1.json).
TEST: seeds 501-548, arms random, ucb, freqrule, learned (online ridge as pilot 1), transfer, transfer_shuf; 300 audit tasks per seed never used in decisions; run once.
Primary contrast (only one): transfer - random, mean audit solved, paired over seeds, 95% bootstrap CI. 'Advantage in this domain' iff CI lower bound > 0.
Otherwise: no advantage demonstrated. Secondary (descriptive, no multiplicity adjustment): transfer vs freqrule, vs transfer_shuf, vs learned, freqrule vs random.
Chronology is by declaration (local file hashes in HASHES.txt), not independently timestamped.

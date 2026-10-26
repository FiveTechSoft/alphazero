# HypothesisEngine pilot 4 - TEST results (seeds 701-748) and accounting
Frozen protocol/code: he_pilot4/ at the public pre-TEST snapshot (PREREG.md, main.go, elig.go, elig_test.go, analyze4.py, run_pilot4.sh, weights_elig.json, train.json, HASHES.txt). Domain: synthetic program search. The selector is a ridge model whose ~12 coefficients were fitted offline on existing labelled edits (statistical training, no new labels, no neural network). Not RSI, not general learning, not discovery of a representation: features, ridge form, eligibility rule and label are programmed.

## Preregistered result (analysis4.txt; mean audit solved of 300, 48 paired seeds 701-748, 3 arms, each arm-seed hard-capped at exactly 6,000,000 decision expansions)
Arm means (AuditSolve; table order is alphabetical, not execution order, which within each seed was learned_elig, learned_elig_shuf, excl_random): excl_random 156.23 | learned_elig 179.50 | learned_elig_shuf 145.48.
PRIMARY learned_elig - excl_random: +23.27, 95% bootstrap CI [16.81, 29.73]. Rule: lower bound > 0, so the criterion is met. In this synthetic benchmark, a selector with ridge coefficients fitted on existing eligible rows chose better than random choice among eligible candidates, for this package of choices (fit subset, tiered policy, one random slot; acceptance ties favour the ridge-chosen edits).

## Secondary (descriptive, no multiplicity adjustment, no causal reading)
learned_elig - learned_elig_shuf: +34.02 [27.62, 40.40]. learned_elig_shuf - excl_random: -10.75 [-15.00, -6.94].
Per-arm diagnostics (alphabetical order; means over 48 seeds):
| arm | rounds | accepts | evals | final library size | fallback rounds | mean wall ms | mean ops |
|---|---|---|---|---|---|---|---|
| excl_random | 40.48 | 8.27 | 122.06 | 16.27 | 0.0 | 1241.94 | 744761.5 |
| learned_elig | 46.02 | 10.58 | 138.75 | 18.58 | 0.0 | 1223.65 | 871319.5 |
| learned_elig_shuf | 38.79 | 6.81 | 117.25 | 14.81 | 0.0 | 1232.83 | 714509.0 |
Fallback (fewer than 3 eligible candidates) occurred in no round of any arm, so the ridge extrapolation to deprioritised candidates did not come into play. Wall time is on one shared core and noisy.

## Limits
One synthetic world family, one budget, audit inside the same family. The package of choices is not separated into its parts: no arm isolates the fit subset, the tiered policy or the random slot, and the result does not show which of them matters. Labels come from the collect regime of pilot 2 (overlap with the regime of use not measured). The permuted-label control differs from excl_random descriptively only. The ridge was fitted on the labels of pilot 2 (seeds 201-232); the TEST worlds 701-748 are new and were not used in the fit, so the contrast is out of sample within the same synthetic family. It says nothing about other families, other budgets, other regimes, or about the separate contribution of the components. The weights (including +3.377 on the corpus-frequency feature) were computed and seen before the freeze. The result does not show that the preference was "discovered": the feature set contains corpus frequency by construction.

## Cost accounting
- TEST: 144 arm-seed runs x 6,000,000 = 864,000,000 decision expansions (Spent = 6,000,000 verified in every row by run_pilot4.sh and analyze4.py).
- Preparation before the freeze: ridge fits on existing data (milliseconds, no decision expansions); unit tests on artificial data (zero expansions); one load check of the learned arms (seed 951, budget 20,000 each = 40,000 expansions). The 192M expansions that produced the labels were spent in pilot 2 and are not counted in the pilot 4 budget.
- The 192M historical label expansions used to obtain the weights do not disappear: they remain a cost of the learned arms. This pilot does not compare total cost at equal compute against excl_random, which needs no labels. The op counter (a proxy for feature and ridge inner-loop work, not total compute) averaged 871,319.5 for learned_elig against 744,761.5 for excl_random, about 17% higher.
- Not counted in those numbers: audit evaluation (300 tasks per seed), pool generation, features, compilation, unit tests run by the driver, wall time. Mean wall time per arm-seed run was about 1.22-1.24 s; the whole TEST took about 3 minutes on one local core. No server, no paid API.
- Interruptions: reported execution details (not independently verified): one continuous execution of run_pilot4.sh, no interruption and no retries. The four complete chunk logs are consistent with this; they do not prove the absence of retries.

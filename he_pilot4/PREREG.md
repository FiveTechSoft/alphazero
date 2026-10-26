# HypothesisEngine pilot 4 - protocol DRAFT v1 (not frozen; TEST seeds 701-748 NOT run)
Domain: the same synthetic program-search domain and code as pilots 2 and 3. Not neural, not RSI, not general learning, and not discovery of a representation (a ridge is fitted to existing labels, which is statistical training, with no new labels): the features, the ridge form, the eligibility rule and the label are programmed; only the ~12 ridge coefficients are fitted to existing data.

## Single question
Does a selector whose ridge coefficients were fitted offline, on the existing labelled edits that were eligible for selection (not a remove and not redundant), choose better than random choice among eligible candidates, at the same hard budget of decision expansions, measured by mean audit-solved on new seeds?
This tests a PACKAGE of choices (fit subset + tiered policy + one random slot), not any one of them in isolation.

## Why this question (declared)
In pilot 2 the offline ridge was fitted on all 4848 labelled edits and its coefficient on the corpus-frequency feature was negative (-7.86) with interaction terms (+6.54, -6.56); in pilot 3 ordering by frequency beat random ordering within the eligible set (+10.15, CI [3.75, 16.56]). Remove edits have large negative labels (mean -3.18) and, by construction, high frequency. Hypothesis, not demonstrated: fitting on all rows mixed the remove effect with the frequency effect. Fitting only on eligible rows gives, for the same ridge and features, a coefficient on frequency of +3.377 with interactions +0.926 and +1.642 (descriptive, in-sample; no out-of-sample performance was evaluated, and a positive coefficient on one feature with interactions does not imply a positive marginal preference for every candidate).

## Data used for fitting (existing, no new labels)
train.json (pilot 2 artifact, sha256 a3c0717034dd486b9a21917b39371344acbc7531db1475f1269453276e87356c, matches the pilot 2 manifest): 4848 rows; eligible rows (X[2]==0 and X[5]==0): 4017 (783 remove rows and 48 redundant non-remove rows excluded; no row is both). So the eligible set is in practice "non-remove" (redundant rows are rare in this data). The labels come from the collect regime of pilot 2 (24 candidates evaluated per round, a random improving edit accepted, libraries of 6 to 16 actions), which differs from the regime in use; the overlap is not measured. The 192M expansions that produced those labels were spent in pilot 2 and are not part of the pilot 4 test budget.
Fit: ridge lambda = 1 on all 12 coefficients including the bias, the same Ridge.fit as pilot 2 (mode fitelig); label-permuted control fitted on the same eligible rows with permutation seed 424243. Cross-checked against an independent numpy solution (max abs difference 1.09e-11) and a Cholesky reference in unit tests.

## Arms (3)
1. learned_elig - frozen ridge weights (weights_elig.json, key learned_elig).
2. learned_elig_shuf - the same pipeline with the label-permuted weights (information control).
3. excl_random - unchanged from pilot 3: deprioritised candidates (remove, or expansion contained in a library action) get score -1 + noise; the others get random scores.
Pool, batches, acceptance rule, pickM = 3 and audit code paths are unchanged. freqrule is not an arm here (its pilot 3 difference from excl_random was measured on other seeds and is not paired with these).

## Selection rule of the learned arms (complete specification)
For each round: pool of 24 edits; idx = ar.Perm(24) (drawn for every arm before selection); then one ar.Float64() per pool item, in pool order, as tie-break noise (exactly 24 draws; verified by a unit test).
Eligible = not (remove or redundant), the same definition freqrule and excl_random use (feature 5 is 1 iff the expansion is contained in an existing action's key; keys have one character per primitive, so containment is contiguous-subsequence containment).
Order = eligible first (strict priority, independent of the scale of the ridge score), then ridge score descending, then noise descending, then lower index.
Choices 1 and 2 = the first two of that order. Choice 3 (random slot) = the first index in idx that is not already chosen and is eligible; if no eligible candidate remains, the first index in idx not already chosen.
Fallback: if fewer than 3 candidates are eligible, some deprioritised candidate is chosen (in the ridge choices if fewer than 2 are eligible, in the random slot if exactly 2). The ridge scores of deprioritised candidates are extrapolations: the fit never saw those rows. Result.Fallback counts such rounds per arm (counted at selection, including a final truncated round); results are reported with and without them described, and not extrapolated.
excl_random has no ridge slots: three random picks among eligible candidates (same fallback behaviour as pilot 3). The learned arms therefore differ from it in two of the three picks being ridge-ranked and in the exact way the third pick is drawn.

## Acceptance order (documented property of the shared code)
After the chosen edits are evaluated in the order of the chosen list, the accepted edit is the first one with the strictly largest positive gain (the code keeps an edit only if its gain is greater than the best so far, bestD starting at 0). A tie in positive gain therefore favours the earlier-evaluated edit. In the learned arms the order is [ridge choice 1, ridge choice 2, random slot], so ties favour the ridge-chosen edits over the slot. In excl_random the order is the order of its random scores. This is part of the package of choices; its size of effect is not measured and not claimed.

## Seeds, budget, cost
Seeds 701-748 (48), fixed now for all three arms. Not used by any run known to the preparer (pilot 1: 1-8, 101-148; pilot 2: 201-232 fit/collect, 401-408 dev, 501-548 test; pilot 3: 601-648 test; smoke and load checks: 901-903, 951).
Hard cap 6,000,000 decision expansions PER SEED PER ARM (a round cut by the cap is discarded, its spend counts). Planned decision expansions: 3 x 48 x 6,000,000 = 864,000,000 (144 arm-seed runs). Not in that number, to be recorded: audit evaluation (300 tasks/seed), pool generation, features, ridge fit, compilation, tests, wall time and the op counter. No new label generation and no neural training. The ridge fit is statistical training on existing labels: its cost (milliseconds, no decision expansions) is recorded above.
Preparation cost already incurred (no TEST seeds): ridge fits on existing data (milliseconds, no expansions); unit tests on artificial data (zero expansions); one load check of the two learned arms on seed 951 with budget 20,000 each = 40,000 expansions (zero complete rounds, so it only checks that weights load and the arms run; it is not evidence). Total preparation expansions: 40,000.
No power calculation is claimed; 48 seeds repeats pilots 2 and 3.

## Analysis (analyze4.py, once)
Validation first (3 arms x 48 seeds 701-748 once each, Spent = 6,000,000, keys, hash entries and hashes of all required files); on failure preserve everything and stop for review; no repair, no algorithm change, no silent rerun.
Primary and only preregistered contrast: learned_elig - excl_random, mean AuditSolve paired by seed, 95% bootstrap CI (10000 resamples, rng seed 1).
- LB > 0: the learned selector beat random choice among eligible candidates in this domain, for this package of choices.
- CI contains 0: no advantage demonstrated. Not equivalence, no cause.
- UB < 0: the learned selector is below excl_random.
Secondary, descriptive, no multiplicity adjustment, no causal reading: learned_elig - learned_elig_shuf; learned_elig_shuf - excl_random; per-arm rounds, accepts, evals, library size, Fallback, wall time and ops. Arm order in tables: alphabetical.

## Known confounds and limits
(a) The package: fit subset, tiered policy and the random slot's source differ from pilot 2's transfer arm at once; there is no arm that separates them. (b) Label regime mismatch (collect vs use), not quantified. (c) Labels are immediate gains on 40-task batches, in-sample ridge, no out-of-sample check. (d) One synthetic world family, one budget, audit in the same family. (e) Selection streams are seeded by seed*1000+99+7919*len(arm name): the three arms (names of length 12, 17 and 11) have different streams; batches, pools and audit do not use them. (f) The weights were computed and seen before the freeze. A positive result would show only that coefficients fitted on existing eligible rows give a selection rule better than random choice within the eligible set in this benchmark. A negative or inconclusive result would not show that the preference is unlearnable.

## Environment and run policy
go version go1.26.6 linux/amd64 (go.mod says go 1.21); Python 3.10.12. A discrepancy: stop and report. Before running: sha256sum -c HASHES.txt, then run_pilot4.sh (explicit commands: environment check, manifest check, initial check that no output exists (including the binary he and he.sha256, so nothing is overwritten), go vet, go test, build, binary hash recorded with sha256sum he | tee he.sha256, four chunks 701-712, 713-724, 725-736, 737-748 each with a per-chunk check that none of its files exists and a validation of 3 arms x 12 seeds once each with Spent 6,000,000 before the next chunk, then analyze4.py). Interruption or failure: preserve log and output, record the attempt with a unique name, stop and report before relaunching.
Freeze (public commit + manifest hash) before any seed 701-748. Chronology is by declaration and commit order, not independent timestamps.

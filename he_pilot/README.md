# HypothesisEngine pilot 1: learned edit selector vs controls (synthetic program-search domain)

IMPORTANT caveats (read first): (1) budgets are approximate and NOT equal in total cost: ~6.00-6.18M expansions for most arms, exhaustive up to 7.04M (+17%, mean +8%), and ridge/pool generation/features were not counted or measured; (2) the evaluation criteria were declared in PREREG.md before the final seeds, a declaration that is not independently verified; (3) candidate pools use the same seeded generator stream but diverge across arms once libraries diverge; (4) one earlier launch was interrupted by the environment, not used, counted as one attempt.

Files: main.go (solver, edit catalog, arms), PREREG.md (design/criteria frozen by declaration, not independently audited), analyze.py, final.json/final*.log
(confirmatory, seeds 101-148, run once), dev_v1/dev_v2 (DEV seeds 1-8, debugging only, NOT evidence).

Status of runs
- DEV seeds 1-8, v1: learned (pure greedy ridge) was worst; it locked onto one edit type. v2: added one random
  exploration slot and a UCB control (replacing a type-mean control). Learned still below most controls on DEV.
  No further tuning after that.
- CONFIRMATORY seeds 101-148 (48 seeds, 6 arms, target budget ~6M solver expansions per arm (selection evaluations included; real spend differs by arm, see COST_ACCOUNTING.md; this is not equal total cost), audit 300 tasks
  per seed never used in decisions): see final_analysis.txt. One earlier launch was interrupted by the environment (not used, counted as an attempt); the same deterministic program was rerun in four chunks; seed 101 reproduced identically. See COST_ACCOUNTING.md.

Result (mean audit solved of 300): freqrule 163.5, exhaustive 154.8, random 151.4, learned_shuffled 149.5,
learned 148.0, ucb 147.9. Paired learned minus: random -3.4 [-8.5,+1.3]; ucb +0.1 [-4.5,+4.5];
learned_shuffled -1.5 [-7.4,+4.4]; exhaustive -6.8 [-12.1,-1.6]; freqrule -15.6 [-21.7,-9.2].
Conclusion: pre-declared criterion NOT met. The learned selector did not beat random, simple UCB or its own
label-shuffled control. A hand rule (frequency of the candidate in the system's own solutions) had the highest mean
audit score (an observed difference, not adjusted for multiplicity across contrasts); a hypothesis for later testing is that a
simple hand-written signal carries most of the usable information in this setup, not a finding.
Limits: synthetic domain written by us, one budget and one hyperparameter setting, ridge selector only, small
history (~100-200 labelled edits per run), noisy 40-task selection batches (winner's curse), efficiency vs
expressivity not separated. Not a neural or RSI result.

Details: budgets are not strictly equal (real spent per arm in COST_ACCOUNTING.md; exhaustive up to +17%). Candidate pools share the same seeded generator stream per seed but diverge across arms once libraries diverge, so pools are not identical across arms. Ridge/pool generation are not counted as expansions. No superiority claim at equal cost for the exhaustive arm.

NOTE (added at export, original text below unchanged): this is a design/criteria declaration written before the final seeds were run; it is not independently verified or externally timestamped. Budgets turned out approximate, not equal in total cost; candidate pools diverge across arms as libraries diverge; one launch was interrupted and not used (see README/COST_ACCOUNTING).

# HypothesisEngine pilot: preregistration (frozen before final seeds 101-148 were run)

Question: at equal total search cost (selection included), does a selector that learns from the history
of edits (version features, edit, measured effect) pick better library edits than random and simple controls?

Domain: enumerative program search (best-first over behaviour states, per-task cap 1500 expansions, depth 8) over
a library of actions (primitive sequences). Hand-written typed edit catalog: add macro of 2 lib actions, add
random 3-primitive macro, remove an action, add a reservoir primitive. Tasks: compositions of 1-3 hidden motifs
(per-seed world). No neural model, no LLM, no pretrained weights. Capability is in the catalog/solver design
(hand-programmed); nothing here is a RSI claim.

Arms (same seed => same world, same task stream, same candidate pools; budget 6,000,000 expansions each,
counting baseline evals and every candidate eval): random (3 of 24 per round), ucb (per edit type), freqrule
(hand rule: n-gram frequency in own solutions), learned (ridge on history, 2 greedy + 1 random slot, first 3 rounds random),
learned_shuffled (same, labels permuted: control for "any learning"), exhaustive (all 24 per round; best-of-N control).
Acceptance: best candidate with positive solved-delta on a fresh 40-task batch each round (renewable selection set).
Audit: 300 tasks from a separate stream, never used in any decision, evaluated once per final library.

Seeds: 1-8 were DEV (debugging; two design fixes after dev: UCB arm replacing type-mean, exploration slot in learned).
FINAL: seeds 101-148, run once, no re-run, no tuning after.
Primary: paired audit-solved difference learned minus {random, ucb, freqrule, learned_shuffled, exhaustive},
mean and bootstrap 95% CI over 48 seeds. "Learned selection helps" only if CI > 0 versus random AND ucb AND
learned_shuffled. Efficiency vs expressivity is not separated in this pilot (reported as unresolved).
Known limits: synthetic domain written by us; single budget; one hyperparameter setting; audit shares task distribution.

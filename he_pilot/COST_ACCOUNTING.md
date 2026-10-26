# Cost accounting (measured, not assumed identical to wall-clock)
Budget unit = solver node expansions (one action applied to one search node), summed over every task evaluation
used for decisions: per round, 1 baseline evaluation of the current library + each candidate evaluation on the
same 40-task batch (all arms). A round starts only while spent < 6,000,000, so arms overshoot by up to one round.
Measured final spent (48 seeds): random 6.07M, ucb 6.07M, freqrule 6.07M, learned 6.08M, learned_shuffled 6.07M
(max 6.18M); exhaustive mean 6.51M (max 7.04M, up to ~17% over the target). The extra spend may favor that arm; no attribution at equal cost is made.
NOT counted in the budget: ridge fit, edit-pool generation, n-gram/frequency features, and the final audit
(mean ~0.25-0.28M expansions per arm, evaluated once). Selector overhead (ridge, pool generation, features) was not measured in either expansions or wall time; wall time was not used as the budget. Discarded work: candidate evaluations that were not accepted
are counted (they are the cost of selection). Candidate evaluations per run: ~120 (random/ucb/learned), ~129 freqrule, ~164 exhaustive.

Real spent expansions per arm over the 48 final seeds (budget target 6,000,000; the budgets are NOT strictly equal):
| arm | spent min | mean | max |
|---|---|---|---|
| exhaustive | 6011208 | 6507049 | 7037227 |
| freqrule | 6000239 | 6072871 | 6156123 |
| learned | 6003659 | 6075131 | 6182164 |
| learned_shuffled | 6004966 | 6070329 | 6161238 |
| random | 6000367 | 6074896 | 6174687 |
| ucb | 6001141 | 6066057 | 6143277 |

Because exhaustive spent more (mean +8%, max +17%), no claim of algorithmic superiority or inferiority at equal cost is made for it.

Chronology (declared, not independently audited): the design/criteria file PREREG.md was committed in a local git repo (eb88d36) before any
seed 101-148 run; that commit timestamp is from the local clock and is not an independent timestamp. Treat this as an evaluation frozen by
declaration, not an independently verified preregistration. One earlier full launch was cut by the environment after ~15 seeds; its output was
not used and is kept privately (it counts as 1 attempt). The same deterministic program was rerun in four seed chunks (f1-f4 logs); seed 101 reproduced bit-identically. No re-run after seeing results, no audit queries used for adjustment.
Not counted as expansions: ridge fit, candidate-pool generation, frequency features (cost not measured).

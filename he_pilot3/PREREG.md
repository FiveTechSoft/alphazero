# HypothesisEngine pilot 3 - protocol DRAFT v2 (TEST seeds 601-648 NOT run at freeze)
Domain: same synthetic program-search domain and code as pilot 2 (he_pilot2/main.go). Not neural, not RSI, not general learning.

## Chronology (declared, not independently timestamped)
- 26 Oct 2026 (CET): design proposal B was written first; the code, this draft PREREG and a smoke run on seeds 901-903 were produced afterwards, the same afternoon.
- The design is conditioned by pilot 2 results and by a retrospective exploratory diagnostic of pilot 2 artifacts (exploratory/). It was NOT fixed after the smoke run: the proposal predates the implementation and the smoke.
- The smoke analysis output (seeds 901-903, n=3) was seen before freeze. Declaration, not verifiable: no design change was made after seeing it.
- The fixtures (12M expansions) were run after the first version of this package was assembled.

## Question (single primary)
Within freqrule's selection regime (candidates that are a `remove` edit, or whose expansion is already contained in a library action, are DEPRIORITISED), does ordering the remaining candidates by corpus frequency give a different mean audit-solved than ordering them randomly, at the same hard budget?

## Arms (only the scoring step of the selector differs; pool, batches, acceptance, pickM=3 and audit code paths are unchanged)
1. freqrule - unchanged from pilot 2. 2. excl_random - same deprioritisation set, random order among the rest. 3. random - unchanged (reference for secondaries).
"Deprioritised" = score -1 (+1e-6 noise), not absolute exclusion: with fewer than pickM non-deprioritised candidates, topK fills with deprioritised ones, ordered by the noise term. Fallback rounds are counted per arm (Result.Fallback, counted at selection, including a final truncated round).
Ties: freq is a multiple of 1/len(corpus) (>= 1/300), so the 1e-6 noise only breaks exact freq ties. Result.Ties counts freqrule rounds where the pickM-th chosen candidate ties with the best unchosen non-deprioritised one. In non-evidence smoke on seeds 901-903 this happened in 17/39, 30/57 and 28/52 freqrule rounds. "Ranking by frequency" therefore means: prefer candidates whose expansion appears in the corpus, with substantial random tie-breaking.

## RNG streams (documented, binary not changed)
Selection stream: seed*1000+99+len(arm name)*7919. excl_random (11 chars) therefore has a stream different from freqrule (8) and random (6); batches (sr stream), pools (hash of seed, round, library) and audit (aur stream) do not consume it. excl_random consumes one ar.Float64() per pool item after the same ar.Perm as freqrule. Out of scope for pilot 3: in pilot 2 the arms "freqrule" and "transfer" have equal name length (8), so they shared one selection stream formula; not relevant to the 3 arms here.

## Budget and seeds
Hard cap 6,000,000 decision expansions PER SEED PER ARM (a round cut by the cap is discarded, its spend counts). Seeds 601-648 (48), the same for all 3 arms. Planned complete decision expansions: 3 x 48 x 6M = 864,000,000 (144 arm-seed runs). Not in that number, to be recorded: audit evaluation (300 tasks/seed), pool generation, features, compilation, fixtures, smoke runs, partial/interrupted runs, wall time and op counter. No offline training.
Smoke/fixture expansions already spent (seeds 901-903, not TEST, not evidence): 198M (33 complete arm-seed runs of 6M) in the preparation runs, plus 12M fixture runs, plus builds and the exploratory script.
The seed count 48 repeats pilot 2; it is not derived from a power or MDE calculation, and no such calculation is claimed.

## Commands, environment and run policy
Environment declared: go version go1.26.6 linux/amd64 (go.mod says go 1.21), Python 3.10.12. A discrepancy: stop and report; do not change it afterwards.
Before running: sha256sum -c HASHES.txt (complete manifest). The manifest hash is reported separately at the freeze.
Exact commands are in run_pilot3.sh (hashed, part of the package): set -euo pipefail; environment check; sha256sum -c; go build -o he .; then 4 chunks 601-612, 613-624, 625-636, 637-648 (./he -lo L -hi H -arms freqrule,excl_random,random -out chunk_L_H.json | tee chunk_L_H.log), preceded by an initial check that none of the 8 chunk JSON/log files nor analysis3.txt exists (before building or running), and each chunk again preceded by an explicit if [ -e json ] || [ -e log ] ... exit 1 (noclobber does not protect tee or os.Create; stderr is captured with 2>&1 before tee), and each followed by a validation that the JSON has exactly 3 arms x 12 seeds of the chunk, once each, Spent = 6,000,000, before the next chunk starts; finally analyze3.py. The last log line is not used as proof of completion.
Interruption or failure: preserve log and output, record the attempt with a unique name, stop and report before any relaunch. No silent repeat, no algorithm change, no mixing of partial outputs.
No weights file is needed by these arms.

## Analysis (analyze3.py, once)
Validation first: exactly 3 arms x 48 seeds 601-648, each (arm, seed) once, Spent = 6,000,000, all required keys present, HASHES.txt has entries for main.go, go.mod, analyze3.py and PREREG.md and their hashes match. On any failure: preserve all files and stop for review; no repair, no algorithm change after seeing TEST, no silent rerun.
Primary and only preregistered contrast: freqrule - excl_random, mean AuditSolve paired by seed, 95% bootstrap CI (10000 resamples, rng seed 1).
- LB > 0: ordering by frequency is associated with higher audit-solved than random ordering among non-deprioritised candidates in this domain (includes trajectory feedback).
- CI contains 0: no advantage demonstrated. Not equivalence, no cause.
- UB < 0: freqrule below excl_random. Says nothing about freqrule against random in general.
Secondary, descriptive, no multiplicity adjustment, no causal reading: excl_random - random, freqrule - random, per-arm rounds/accepts/evals/library/Fallback/Ties/Spent/wall/ops. Remove vs redundancy cannot be separated by this design (not factorial).

## Known confounds
Corpus/trajectory feedback; heavy tie-breaking; one synthetic world family, one budget, audit in the same family; separate selection streams per arm; 1e-6-noise ordering of deprioritised candidates; design conditioned by pilot 2 and a seen smoke.

## Process commitments
Freeze (public commit + hashes of main.go, go.mod, analyze3.py, PREREG.md, README.md and the other listed files) before any seed 601-648. Interruptions and partial compute logged from the start. Chronology = declaration plus commit order.
Status at freeze: TEST seeds 601-648 not run.

# Prospective H/C comparison D1

Design only. No experiment implementation/execution. Compare full historical H versus frozen joint contract C, not causal effects of individual rule changes.

32 paired independent training replicas, r=0..31. Same initial weights/moments for both arms, momentszero/t0. Seeds: weights2026110100+10r, train2026110101+10r, exam5=2026110102+10r, exam4=2026110103+10r. Freeze all lists, ASTs, corpus, labels and ordering before either arm trains.

Per replica16train tasks:4each levels2/3/4/5. Common uniform-prior teacher,800expanded nodes per train task. Same64TrainCEcalls in each arm; cycle corpus from start if shorter, stop if empty. Do not generate different self-play data per arm. No tuning, early stopping, best checkpoint selection or additional caps.

Exam16new level5cases primary and8level4secondary per replica,300expanded nodes each, pre/H64/C64. Exclude identities, duplicate/inverse pairs and overlap with ALL train lists. Audit variable-renaming templates without claiming semantic independence. Exam never supplies training or labels. Exam CE is valid only with independently frozen valid labels; train loss does not decide a winner.

Primary D per seed=postC-success-rate minus postH-success-rate on level5. Report all seeds, gains/losses from shared pre, onlyC/onlyH/both/neither, verified paths, nodes including failures, finiteness and separate teacher/train/exam CPU/wall.64calls do not mean equal compute; report both.

Single confirmatory paired sign test across NON-TIED seeds, bilateralalpha.05: W=#Dpositive,L=#Dnegative,m=W+L; p=min(1,2*sum(j0..min(W,L),choose(m,j)*2^-m)); m0givesp1. Consistent observed advantage additionally requires matching majority direction and observed mean magnitude>=5percentagepoints. This does not prove a population mean advantage>=5pp. Secondary metrics descriptive.

Per-seed tie=same solved count even if cases differ. Non-significance is inconclusive, not equivalence. Practical observed tie: absolute mean<5pp without consistent advantage; equivalence remains unproved. Larger uncertain differences are inconclusive. Report level4regression>5pp and median totalCPUratio>1.20as trade-offs; no automatic replacement. Show contextual paired-mean interval with assumptions, not another winner rule.

32is fixed proposed initial sample size, NOT guaranteed80%power. Prior n16was only2historical seeds and another pre/post level4comparison; no estimate of H-C variance.512cases are grouped in32models, not512independent training replicas. Power needs effect, dispersion and tie-rate assumptions before new results. Do not add seeds until significance. Incomplete study is incomplete, not a subset analysis of successes.

Traceability: immutable source/test hashes, compiler/environment, generator/judge versions, exact fixtures/corpus, full optimizer state, per-case proof/counters/costs, bounded operation and cleanup. No automatic retries or budget increases. Operational admission remains separate from scientific design.

References:
https://www.itl.nist.gov/div898/software/dataplot/refman1/auxillar/signtest.htm
https://www.itl.nist.gov/div898/handbook/prc/section3/prc311.htm

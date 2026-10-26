# lang9: curriculum contrast (D1 then D2 vs D2 only)

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Arms per seed (21, 22): cur = D1 for updates 0-1999 then D2 for 2000-4999, no reset of weights, optimizer or schedule; ctl = D2 for all 5000 updates from the same init. Same model, optimizer and episode stream as lang8. Batched evaluation (batch <= 128) reuses the lang8 frozen eval set.
Outcome: no benefit in either seed (benefit_both false, strong_both false). E_bl exact: seed 21 cur 0.0625 vs ctl 0.0694; seed 22 cur 0.0741 vs ctl 0.0741. In seed 22 the exact-match values of the two arms are identical on every split although the final weight files differ (sha differ); no explanation was investigated. D1 accuracy after the switch phase and retention at the end are in results/runs/result_c9.json (seed 21: TR 0.603 at the switch, 0.144 at the end).
Limits: one curriculum split (2000/3000), two seeds, one model size.
Omitted: launcher smoke tests (smoke, smoke3) and script self-tests, which contain no experiment.

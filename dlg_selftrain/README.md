# dlg_selftrain — status
Experiment in progress (2026-10-26). Design: DESIGN.md.

Measured so far (seed 7001; dev = monitoring sets, test = reserved, evaluated
once per arm at its closeout):
- base adequacy gate: PASSED (dev_seen explicit accuracy 0.5574, n=183,
  threshold 0.30).
- b0 (base only, no self-training rounds): test_seen 0.715, test_novel 0.585,
  test_novel_long 0.537. These are the reference numbers the self arm must
  beat; they are NOT evidence of self-training improvement.
Details: results/base_b0_seed7001.json. Arms sup/self/selfuf and seed 7002
pending. Nothing here is a claimed improvement yet.

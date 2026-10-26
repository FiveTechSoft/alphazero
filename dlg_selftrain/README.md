# dlg_selftrain — status
Experiment in progress (2026-10-26). Design: DESIGN.md.

Measured so far (seed 7001; dev = monitoring sets, test = reserved, evaluated
once per arm at its closeout):
- base adequacy gate: PASSED (dev_seen explicit accuracy 0.5574, n=183,
  threshold 0.30).
- b0 (base only, no self-training rounds): test_seen 0.715, test_novel 0.585,
  test_novel_long 0.537. These are the reference numbers the self arm must
  beat; they are NOT evidence of self-training improvement.
Details: results/base_b0_seed7001.json.
- sup (free-supervision reference, 3 rounds gold answers, seed 7001):
  test_seen 0.772 (+0.057 vs b0), test_novel 0.575 (-0.010),
  test_novel_long 0.527 (-0.010): more supervised rounds improved SEEN
  accuracy but not novel-compositional accuracy on this seed.
  Details: results/sup_seed7001.json.
- self (verified own answers, 3 rounds, seed 7001): test_seen 0.800,
  test_novel 0.608, test_novel_long 0.523. Point differences vs b0 on this
  seed: +0.085 seen, +0.023 novel, -0.014 novel_long; vs sup: +0.028 seen,
  +0.033 novel, -0.004 novel_long. One seed, no uncertainty estimate yet;
  H1 (self > b0 on test_novel on BOTH seeds) is open until seed 7002.
  Details: results/self_seed7001.json.
Arm selfuf and seed 7002 pending.

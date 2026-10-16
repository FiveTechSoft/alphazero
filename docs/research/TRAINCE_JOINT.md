# Joint TrainCE contract

Prepared source, NOT compiled/tested. Accumulate W1/W2 gradients before writes; dh uses oldW2; L2W1 once on oldW1. t once/sample and one m/v/q update per parameter. B1 and moments fixed; B2/mb2/vb2 exactly preserved. Same betas.9/.999,eps1e-8,caller lr. No architecture/features/solver changes.
Return value deliberately remains approximate preupdate CE without L2 for compatibility, not the exact objective CE+L2. Reference objective is logsumexp(scores)-scoretarget+lambda/2||W1||²,lambda1e-4,no division by actioncount.
Validation: selector ^(TestTrainCEJointContract|TestTrainCEJointFourCases)$; env CE_CONTRACT_AUDIT=1,CE_JOINT_AUDIT=1. First is original unmodified numerical test against this branch C. Extra test checks cold/warm x012/201 with frozen fixture/tolerance and full-state finiteness. NO observed PASS. Historical replay test intentionally excluded from C selector.

Test names/gates refer to the complete preserved validation suite, intentionally not included in this public source branch. No validation result is claimed.

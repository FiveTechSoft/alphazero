# Minimal Expanded fix

Prepared, NOT compiled/tested. Static bug: expand never set Expanded=true, selection skipped and visits stayedzero. Set flag after initializing node before both returns. No full motor rewrite.
TTT value from side-to-move requires alternating sign. Optional MCTSValuePerspective.AlternatingTurns()false allows future single-player fixed-perspective values without negation; existing games defaulttrue. No algebraMCTS integration.
Prepared MCTS_REGRESSION_AUDIT tests: TestMCTSExpandedRegression requires512rootvisits,exploration,winningmove8 distinct from firstlegal3. Original winningmove2fixture unchanged but could pass first-action bug. TestMCTSSinglePlayerPerspective uses fixed-reward toy,not algebra. NO observed PASS.
Children slice,TotalNcache,pathpreallocation NOT applied: benchmark evidence required, unnecessary for functional fix. No speedup claim. Terminal/no-action assumptions unchanged.

Test names/gates refer to the complete preserved validation suite, intentionally not included in this public source branch. No validation result is claimed.

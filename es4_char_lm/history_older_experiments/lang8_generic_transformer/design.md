# lang8: generic transformer, same task and serialization

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Decoder-only causal transformer, 2 layers, d=64, 4 heads, 103,104 parameters, tied head over a 20-token vocabulary, GPT-2-style init, AdamW lr 1e-3 (warmup 200, clip 1.0), batch 32, 5000 updates (reduced from a 6000 target before training for a time budget; a negative may reflect under-training). Input is [SUP w c ...][LAY ...][Q ...][OUT]; output = [cell_src, cell_dst, STOP], free greedy decoding. No programmed pointer or routing.
Part 1: arms full / nocue, seeds 21, 22: not met in either seed (E3 free-exact 0.078 and 0.074; the best baseline is about 0.42 higher (vs_best_baseline -0.42 in both seeds)).
Part 2 (diagnostics D1, D2, no cues): D1 = colour tokens in the query, one contextual hop; D2 = word tokens, two hops. D1 reaches TR exact 0.946 (seed 21) and 1.0 (seed 22) with held-out-layout exact 0.62 and 0.74. D2 stays at about 0.06-0.09 on all splits. Counterfactual layout/support edits are in res_*.json.
The venv used to run these is not included.

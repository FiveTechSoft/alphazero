# lang10: single-hop in-context word->colour retrieval

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Same generic transformer. Prompt: support of 4 word/colour pairs, no layout; target = [colour(a), colour(b), STOP]. Arms: ctx = real support; ctl = support colours drawn from an independent binding (no information). 5000 updates, seeds 21, 22. Splits: TR, E_b (held-out bindings), E_s (reserved query strings), E_bs.
Outcome verdict recorded: BINDINGS_SEEN_ONLY. ctx TR 0.919 / 1.0; ctx E_b 0.538 / 0.690; ctx E_s 0.935 / 1.0; ctx E_bs 0.575 / 0.657; no-context E_b about 0.05-0.06; ctl TR about 0.09-0.11. The criterion for retrieval (E_b >= 0.90) was not met; counterfactual support edits change the prediction in 94-97% of cases and about 57-67% become correct.

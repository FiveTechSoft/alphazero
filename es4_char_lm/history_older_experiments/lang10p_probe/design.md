# lang10p: paired-context probe on the two final lang10 ctx models

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Evaluation only. 3456 paired prompts per context: 72 binding pairs (seen-seen 41, held-seen 26, held-held 5) x 2 ordered queries x 24 support orders; gold identical in both contexts, only the two non-queried support colours differ. Unit = the binding pair. Categories: same and correct, same and wrong, differ with one correct, differ and both wrong. Results and raw predictions are in p10_probe/.

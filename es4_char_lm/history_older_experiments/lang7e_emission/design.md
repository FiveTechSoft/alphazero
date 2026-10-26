# lang7e: learned autoregressive emission of [src cell, dst cell, STOP]

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Same task and splits as lang7. The model emits three tokens; evaluation is free greedy decoding over a 16-token vocabulary with the model's own previous token. The content pointer is a programmed scaffold; the emitted cell id is close to renaming an output. Arms: full / nocue / noprev; seeds 21, 22; 6,000 updates.
Outcome: criterion not met (advance false). TR is 1.0 in both seeds. E3 free-exact: 1.0 (seed 21), 0.853 (seed 22); E5: 0.808 (seed 21), 0.382 (seed 22) with invalid outputs present. See results/out/result.json.

Record edit: the original script wrote a hash of the design-note text file into the result (field `prereg_sha`). That file is not published, so the line was removed from the script and the field from results/out/result.json (see CODE_CHANGES.diff). Nothing else in the result changed.

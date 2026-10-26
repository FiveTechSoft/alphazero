# calib2: resource measurement, no learning (not the ES4 model; no weights)

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Pre-LN decoder, vocab 8192, context 256, d 256, 6 layers, 4 heads, 6,901,760 parameters, batch 2x256, synthetic random token ids, fp32, 1 thread. Per-step wall and peak memory after each step are in out/res.json (step time about 0.25-0.35 s, peak memory above 560 MiB in the first steps). A first attempt with batch 8 was stopped by the memory limit and is documented only in the history of this record, no output kept.

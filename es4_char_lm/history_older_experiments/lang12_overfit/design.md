# lang12: memorisation diagnostic

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

The generic transformer is trained on one fixed batch of 32 episodes for 1000 updates (AdamW lr 1e-3, warmup 200, clip 1.0), seeds 21, 22.
Outcome: exact match after training 1.0 in both seeds (before: 0.0); final loss 5.0e-4 and 4.8e-4. The model and optimizer can fit a fixed batch; this says nothing about generalisation. A backup copy of this directory was byte-compared; it is not duplicated here.

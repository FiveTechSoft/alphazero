# lang5: numpy model, 4 support items, layouts, pair composition

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Task: 24 word->colour bindings (18 train / 6 new), 24 cell layouts, 12 ordered word pairs (8 seen / 4 reserved). Support: 4 seen pairs with their own layouts; query: a new layout. Objective: cells of source and destination colours.
Model: hand-written numpy model (word embeddings, concat-MLP item encoder H=64 d=16, dot-product attention over 4 support items, 2 heads), Adam lr 3e-3, batch 32, seeds 21 and 22.
Parts: step1 = memorisation sanity on one fixed batch (2500 updates; final joint accuracy 1.0, loss 2.4e-5); pilot = arms full / empty support / permuted support labels, 2500 updates, criterion on reserved pairs x new bindings; inv = three follow-up experiments (A query-only, B longer full-contract, C single-layout lookup).
Outcome (pilot): criterion not met in either seed (advance false; see results/pilot/out/result.json). The inv outcome is in results/inv/out/result.json (fields A_pass, A_comp, B_ctx_use, B_transfer, C_pass).
Exact no-context bounds per split were computed and are in the results (E3 0.2917, E1 0.3125, E2 0.1111, TR 0.1111).

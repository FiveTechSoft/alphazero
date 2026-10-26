# lang7: roles from query cue tokens

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

The query is 5 tokens with cue tokens DE (source) and A (destination); two orders per word pair, train uses one order per unordered pair, the other order is reserved. Programmed: previous-token mixing, role-vector selection, shared support attention and pointer. 1,840 learned parameters. Arms: full / nocue / empty / permuted; seeds 21, 22; 6,000 updates.
Outcome: criterion met in both seeds (E3 and E5 joint 1.0, min group 1.0; advance true).
Operational note: the first launch of this phase failed before any model computation (the launch wrapper could not create its log file because its working directory was wrong). The wrapper was corrected and the phase was run once to completion afterwards. No computation from the failed launch exists and no result was discarded or repeated. The only training results are those of that single run; the setup change was in the launcher, not in the model, data or evaluation.
Caveat: new bindings recombine word->colour facts already seen in training.

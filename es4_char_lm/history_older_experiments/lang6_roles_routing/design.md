# lang6: programmed role routing, memory read and content pointer

Design reconstructed retrospectively from the original run records (design notes, code, results). It is a restatement written after the runs, not a contemporaneous pre-registration. Results are in results/. All tasks are small synthetic in-context binding tasks (word -> colour -> cell); none is about natural language, speech, or self-improvement.

Change of problem: support = 4 word->colour items; query = (word1, word2, layout over 4 cells). The model has programmed key-value attention, a content pointer over layout cells, and fixed role routing (first word = source, second = destination); roles are NOT learned. 1,184 learned parameters. Arms: full / empty / permuted support colours; seeds 21, 22; 10,000 updates; Adam lr 3e-3.
Outcome: criterion met in both seeds (E3 joint 1.0, min group 1.0; advance true). The no-context bound on E3 is 0.2917. This shows binding recovery plus colour localisation with the programmed pointer, memory and routing; it does not show role learning.

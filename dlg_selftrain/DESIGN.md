# Self-training on state-tracking dialogues — experiment design (v2)
Dates: design fixed 2026-10-26 06:55 CET (v1), corrections 07:03 CET (v2), before
any training result of the comparison arms. Status: base model trained; arm
results pending. This document restates the protocol and its correction history.

## Question
Can a small character-level transformer trained from scratch improve on held-out
state-tracking dialogues by generating its own answers, keeping only
programmatically verified-correct ones, and retraining (STaR-style) — and how
does that compare to spending the same gradient steps on free ground-truth data?

This is not a claim of recursive self-improvement: the improvement mechanism
(generate -> verify -> retrain), the curriculum, and all hyperparameters are
fixed in advance. The model only learns answers in its weights. Two seeds:
exploratory, not a robustness claim.

## Task
Spanish micro-dialogues, objects x colors with gender agreement
("la caja es roja. ahora es azul. de que color es la caja? azul.").
Turn types: statement, explicit update, pronoun update, explicit question,
pronoun question ("de que color es ahora?"). Verifier: an independent parser
(gen.check) reconstructs the world state from raw text and computes the exact
expected answer string. Cross-validated against the generator's bookkeeping on
4000 dialogues (6,993 questions): 0 mismatches.
Compositional split: 6 (object,color) pairs held out of ALL training paths
(pelota-verde, libro-blanco, taza-gris, coche-amarillo, mesa-morado,
globo-negro); every object and every color still appears in training with other
partners.

## Model / training
Char-level GPT from scratch: D=128, L=4, H=4, CTX=352, 841,600 params,
vocab 25 chars, tied embeddings, AdamW lr 1e-3 (warmup 50), clip 1.0,
batch 16 windows of 352. Same D0 corpus (6000 dialogues, seen pairs only,
seed 5000, sha256 7c55384dfe9dab31421a1c6bedba0a4be3d8dccf15de5dbbe6315914b947114d)
for all arms and seeds. This model is new for this experiment; it is not a
continuation of any previous model's weights.

## Arms (per base seed; all continue from the SAME frozen base model)
- b0   : base only (1500 steps), no rounds. Measures the delta of the rounds;
         it is the base measured before rounds, not an equal-budget arm.
- sup  : +3 rounds x 500 steps on round prompts WITH ground-truth answers
         (free-supervision reference, same prompts as the self arms).
- self : +3 rounds x 500 steps on the model's OWN sampled answers
         (800 prompts/round, K=4 samples, temp 0.9, seeded; keep only
         parser-verified correct, cap 2/prompt) mixed 50/50 with D0 replay.
- selfuf: identical to self but keeps the FIRST sample per prompt WITHOUT
         verification (controls for "training on own outputs" alone).
Same gradient-step budget (1500 + 3x500) for sup/self/selfuf. Generation and
verification compute of self/selfuf is extra and measured (samples drawn,
acceptance, seconds per stage). "Same budget" means gradient steps, not total
cost. Round prompts are identical across arms within a seed (prompt seed
6000+round), drawn from the seen-pairs distribution only (v2).

## Metrics
Exact-match answer accuracy (greedy decode).
- dev_seen / dev_novel (300 prompts, seeds 91001/91002): monitored after base
  and each round. Never used for selection or tuning.
- test_seen / test_novel (400, seeds 92001/92002) and test_novel_long (300,
  longer dialogues, seed 92003): evaluated ONCE per arm, at the arm's end.
- Eval sets v2: seen sets contain zero held-out pairs anywhere in the dialogue;
  novel sets keep only questions whose resolved asked pair is held out,
  pronoun questions included (referent resolved by the parser); all prompts
  deduplicated within set and against all 10,552 unique D0 question prefixes.
  Instrumented counts: SPLIT_FIXTURE_V2.json. Known residual collision:
  2 prompts shared between dev_novel and test_novel, 1 with test_novel_long;
  reported as a limit, sets not modified again.

## Base adequacy gate (fixed 07:01 CET, before reading any base metric)
If dev_seen explicit-question accuracy (n=183) < 0.30, the base is inadequate
for the bootstrap: stop before the arms, report, treat any amendment as
exploratory and outside the confirmatory comparison. Otherwise proceed.

## Hypotheses / decision rules
H1: self test_novel > b0 test_novel on both seeds. Limit: b0 has fewer gradient
steps; H1 alone does not isolate self-training from "more training" — that is
what the equal-step sup and selfuf arms are for.
H2: self <= sup (free supervision is the reference; gap reported).
H3: self > selfuf (the verifier does real work).
Any outcome including null/negative is reported as-is. No early stopping, no
extra seeds, no uncontrolled relaunches after seeing test results.

## v1 -> v2 correction (07:03 CET)
v1 round prompts allowed held-out pairs (measured: 3.5-6.1% of prompts asked a
held-out pair, 20-24% mentioned one; 1 exact overlap with v1 test_novel; v1
test_seen had 9/400 D0 duplicates and 11/400 pronoun questions with held-out
referents), and the runtime loop verified against generator bookkeeping instead
of the independent parser. v1 supported no claim; no arm ran under v1 and no
test set was evaluated under v1. v2 changes: round prompts seen-only, runtime
labels/verification via gen.check, eval sets rebuilt as above, empty-subset
accuracies reported as null.

## Known limits
n=2 seeds; one synthetic templated task family; self-training data that passes
an exact-match verifier is ground-truth-consistent, so the mechanism tested is
bootstrapped consolidation and coverage via sampling luck, not discovery of new
behavior; excluding held-out pairs from training is a necessary condition for
the compositional test, not proof that improvement comes from transfer alone;
results do not transfer claims to other models, tasks, or conversation.

## Reproducibility
Checkpoints (weights+optimizer+RNG+step) are written atomically every 100-250
steps; resume parity verified bit-exact (straight run vs interrupted+resumed
run, single and mixed data streams). Code: gen.py (generator + independent
parser/checker), lm.py (model/training/eval/checkpointing), run_arm.py (arms).

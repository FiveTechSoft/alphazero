# es4_char_lm: character-level language model on a Spanish Wikipedia sample

Small research artifact. A 6-layer, d=256, 4-head GPT (4,842,496 parameters) over 148 characters, context 256, trained from scratch for 22,500 steps on 5,731,361 characters of Spanish Wikipedia text: 5,719,395 characters of paragraph text from 314 articles, plus the spaces used to join paragraphs (the train file is about 5.9 MB because some characters take 2 bytes). Vocabulary: 147 characters plus one UNK symbol = 148. Each step uses a batch of 2 windows of 256 characters. AdamW (betas 0.9/0.999, no weight decay), 100 warmup steps to learning rate 1e-3, then cosine decay to 1e-4. Settings are in `code/es4_torch.py`.

## What this is NOT
It is not a chat model, not speech, not a general language model, and not a demonstration of self-improvement. It only predicts the next character of Wikipedia-style text. See `results/generation_probe/` (the preregistration there is a public restatement written after the run; the design was fixed before generating): the model's real samples for greetings and simple questions are fluent-looking Spanish fragments that do not answer anything, and greedy decoding falls into loops ("de la construcción de la construcción...").

## Result (descriptive, one seed)
Final step 22,500, mean next-character cross-entropy in nats: validation 1.4004, train 1.3292 on fixed windows. Character n-gram baselines (Witten-Bell, orders 1-5, from train only) on the same validation windows: 1.4004 minus baseline = -1.681 (unigram), -1.011 (bigram), -0.662 (order 3), -0.332 (order 4), -0.125 (order 5). Details and uncertainty: `results/final_res.json`, `results/baselines.json`.
Limits: one seed; text from a topic-biased low-id prefix of the dump; 64 fixed validation windows from 28 documents; cleaning by regular expressions is imperfect; a gap between train and validation does not identify memorization; beating an n-gram table does not show long-context use.

## Contents
- `model/model_g075.pt`: final weights (a torch state_dict; loads with `weights_only=True`). sha256 1dd8d510dea6399504b955ba53341467c151dfc770a119d1a2d74347e0f2744d
- `model/adam_g075.pt.part00`, `.part01`: AdamW state at step 22,500, split in two for upload limits. Join with `cat adam_g075.pt.part00 adam_g075.pt.part01 > adam_g075.pt`; sha256 of the joined file 4890d6d7ac83fb6032b66605560f0bc8760df0a2f4aaa98f53875862bfcbb742. `model/train_state_g075.{json,npz}`: step, numpy and torch RNG states, last evaluation. Resuming exactly also needs the data order seed (14014) and the settings in code/es4_torch.py. Exact resumption was tested only inside the original run (continuity check in `results/t0_res.json`), not from this export.
- `data/`: train/val text as used (`train.jsonl`, `val.jsonl`), the same pages after text cleaning but before the later deduplication and overlap filtering (`raw_train.jsonl`, `raw_val.jsonl`; see Data files), `vocab.json`, `manifest.jsonl`, `prep2_report.json`, `LICENSE_SOURCE.json`.
- `code/`: model and evaluation code (`es4_torch.py`, `es4_common.py`), `recover_eval.py` (reload the model and recompute the fixed validation loss; expected 1.4004431963), `gen_probe.py` (generation probe).
- `results/`: final evaluation, baselines (the 'code' checksums field of the original file was removed), continuity check, 75-generation checksums (`GENERATIONS_PUBLIC.json`: checksum fields only; the original hash chain is not reproduced; the checkpoints themselves are not included in this package), generation probe.

## Recover (verified before publication on a copy of this package, see below)
Requirements: Python with torch 2.9.1 (CPU build) and numpy. Join the Adam parts if you need the optimizer state (`cat model/adam_g075.pt.part00 model/adam_g075.pt.part01 > model/adam_g075.pt`; not needed for the check below).
```
export ES4_ROOT=<path to this folder, ending with />
cd code && python recover_eval.py
```
It checks the data checksums against `data/prep2_report.json`, the model checksum against `results/GENERATIONS_PUBLIC.json`, reloads `model/model_g075.pt` with `torch.load(..., weights_only=True)` and recomputes the fixed 64-window validation loss. Expected: 1.4004431963, per-window difference 0. It reads `model/train_state_g075.json` (no pickle) and does not use test data or training. The training-state RNG arrays are in `model/train_state_g075.npz` (load with `allow_pickle=False`).
Generation probe: `code/gen_probe.py` (the exact script that produced the samples) with `ES4_ROOT` as above, `MODEL_PT=<ES4_ROOT>model/model_g075.pt` and `GEN_ROOT=<folder containing an out/ directory>`.

## Not included, on purpose
Intermediate generations g001-g074, earlier experiments and their checkpoints, the third (test) document split in this first upload (planned for a later upload), process logs. This is not a complete backup of the work.

## Data license and attribution
Text derived from Spanish Wikipedia (dump 2026-10-01, https://dumps.wikimedia.org/eswiki/latest/). Original textual content is under GFDL and CC BY-SA 4.0 (https://dumps.wikimedia.org/legal.html); share-alike and attribution apply to the `data/` files and to derivatives. The full author list is not in the dump; per-document titles and revision ids are in the jsonl files. See `data/LICENSE_SOURCE.json`.

## Change note (v2)
`data/LICENSE_SOURCE.json` and the first (header) line of `data/manifest.jsonl` originally carried a note "Private use only" that was an internal usage note, not a license of the source. It was removed; the Wikipedia license (GFDL + CC BY-SA 4.0) and attribution text above stand. `data/prep2_report.json` still lists the sha256 of the ORIGINAL `manifest.jsonl` (63cb3782...), so that one value no longer matches this file; this is the license-header edit, not corruption. The current sha256 of `data/manifest.jsonl` is in `EXPORT_MANIFEST.json`. train, val and vocab are unchanged and still match `prep2_report.json`.

## Weights and code
The weights are a model trained on the Wikipedia text above; no separate license is declared for them here, and the Wikipedia license terms for the training text are given above without a legal analysis of derived weights. The code follows the license policy of the repository that hosts this folder.

## Wikipedia source
`data/LICENSE_SOURCE.json` records the dump URL, dump date and prefix checksum as written when the data was prepared. The URL points to the mutable "latest" listing, and that date was not re-verified at publication time. Per-document attribution: `data/manifest.jsonl` lists title, revision id and revision URL for every document of every split; the train/val/raw files carry id, title, revid and timestamp.

## Data files (measured)
`raw_train.jsonl` and `raw_val.jsonl` are not the raw dump. They hold the same pages after text cleaning (regular-expression extraction of paragraph text) but before the later filtering steps. `raw_train.jsonl`: the same 314 articles, 12,288 paragraphs vs 12,280 in `train.jsonl` (8 exact duplicate paragraphs removed), 5,720,204 vs 5,719,395 characters. `raw_val.jsonl`: 46 articles and 1,975 paragraphs vs 45 articles and 1,958 paragraphs in `val.jsonl` (paragraphs sharing at least 20% of their 8-word sequences with train were dropped, per `prep2_report.json`; the one article that is absent was not traced). The first 16 MB of the dump file is the only source text.
Cleaning is imperfect: 12 paragraphs in `train.jsonl` still contain a '|' character (residue of table or template markup), and other residue likely remains. No data change was made for this archive.
`test.jsonl` and `raw_test.jsonl` (a third document split of the same prefix) are NOT in this first upload; they are planned for a later upload. They are kept apart from train and validation by document; they were not used for selection, tuning or any reported number. Leakage is not excluded: paragraph overlap with train was measured (see `overlap_hist` in the report) and is not proof of absence.

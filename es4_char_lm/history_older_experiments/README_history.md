# Older experiments (history, small synthetic tasks) and the 3-step continuity test

Scope of this folder: the earlier, small experiments that preceded the character-level language model (folder `es4_char_lm`), kept so the record is complete. They are synthetic in-context tasks (word -> colour -> cell binding and composition) and two resource measurements. They are not about natural-language ability, speech, conversation, or self-improvement, and nothing here shows any of those.

`DATA_CATALOG_2B.json` lists every weight, optimizer and data file (313 files: .pt, .npz, .npy, and the two Adam parts of the 3-step test) with size, sha256, format and expected content (array names, dtypes, shapes; tensor names and shapes).

## Contents
Each experiment folder has `design.md` (design reconstructed afterwards from the records, not a contemporaneous pre-registration), `code/` (comments removed, otherwise the original scripts), `results/` (result JSON, local paths scrubbed), `data_weights/` (initial, final and optimizer weights, predictions, frozen evaluation sets, random-generator states).

| folder | what |
|---|---|
| lang5_binding_numpy | numpy model, 3 parts (sanity, pilot, follow-ups) |
| lang6_roles_routing | programmed pointer and role routing |
| lang7_cues | roles from cue tokens (with an operational note) |
| lang7e_emission | learned emission of the output string |
| lang8_generic_transformer | generic transformer; shared model code gt.py, gi.py; diagnostics D1/D2 |
| lang9_curriculum | curriculum contrast |
| lang10_single_hop, lang10p_probe | single-hop retrieval and a paired-context probe |
| lang11_val_selection | practice chosen from validation failures (validation results only) |
| lang12_overfit | memorisation diagnostic |
| calib2_resource, calib_lm_resource | resource measurements |
| t0_technical_test | 3-step continuity test of the language-model run; technical only |

## Conversions and edits
- Random-generator pickles were converted to `.npz` (keys are the nested path, e.g. `np_rng/1` is the Mersenne keys array). The original pickles are not published.
- `.pt` files are plain state dicts; load with `torch.load(path, weights_only=True)`. All 80 were loaded this way before publishing.
- The Adam state of the 3-step test is split in two parts: `cat adam.pt.part_00 adam.pt.part_01 > adam.pt`.
- Python comments were removed with a tokenizer; no code line was changed. Absolute working-directory paths in 28 scripts were replaced by `EXP_ROOT + relative path` (environment variable `EXP_ROOT`, default `.`), plus a 2-line header; the exact changes are in `CODE_CHANGES.diff` of each experiment folder (original paths shown as `<WORKDIR>`). Scripts import each other by the original directory names (`lang8/gt.py`, `lang9/memguard.py`, `lang11/data11.py`, ...), so rerunning needs a directory laid out under `EXP_ROOT` with those names and the original `out/`, `runs/` subfolders; here code, results and data are separated, so that layout must be rebuilt by hand. No script logic was changed. These scripts were not rerun; they were syntax-checked, and one saved model was re-evaluated as a fixture (see below). Some results JSON fields hold hashes of text files that are not published.

- Fixture check (evaluation only, no training): the saved lang8 `full` seed 21 final weights were loaded with `weights_only=True` and decoded greedily on the TR and E3 evaluation sets; predictions equal the stored predictions exactly (648 and 720 prompts).

- T0 hashes: the hashes inside `t0_technical_test/results/t0_res.json` are tensor-content hashes, not file hashes, so they differ from the manifest sha256 values. A static recomputation of that content hash on the published files did not reproduce the stored values (details in its design.md); no match is claimed.

## Not included, with reason (counts in COVERAGE_2B.json)
- Execution logs, launcher scripts and receipts, smoke tests and self-tests: no experiment content.
- Original pre-registration and design text files: replaced by `design.md`.
- `lang11/sealed` (test set): exists, access mode 000, not opened; only hash and construction rule in `lang11_val_selection/results/manifest11.json`. A used-up test evaluation of these weights is held elsewhere and not part of this folder.
- Backup archives and `lang11_final.tar.gz`: members were compared byte by byte; they contain no file that is not already here, except script self-tests and environment notes.
- Python environment of the lang8 runs.

## Limits
Two seeds per condition, one small model class, evaluation sets built by the same generator as training (new bindings recombine seen word-colour facts), no variance estimates. Gaps between arms are descriptive.

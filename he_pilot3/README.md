Pilot 3 preparation package v2 (DRAFT, unpublished, TEST 601-648 not run).
- main.go: pilot 2 code plus arm excl_random and read-only counters Result.Fallback / Result.Ties. main_pilot2_orig.go.txt = pilot 2 original (diff shows all changes).
- go.mod, run_pilot3.sh (TEST driver with explicit commands, not run), analyze3.py (validating analysis), PREREG.md (draft protocol, commands, chronology), HASHES.txt (sha256 of every other file in this package; it does not hash itself, report its hash separately).
- fixtures/: seeds 901-903 (not TEST, not evidence): original binary vs new binary identical in all fields except WallMs and new counters for random/ucb/learned/freqrule; excl_random and freqrule re-run identical. They were produced with weights.json from pilot 2, which is not part of this package and not needed for the pilot 3 arms.
- exploratory/: retrospective exploratory diagnostic of pilot 2 artifacts, with caveats. Not used to justify design, power or seed count.

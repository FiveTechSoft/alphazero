Pilot 4 preparation package (DRAFT, unpublished, TEST 701-748 not run).
- main.go: pilot 3 code plus arms learned_elig / learned_elig_shuf and mode fitelig. elig.go: selector, tiering and ridge fit on eligible rows. elig_test.go: unit tests on artificial data (reference implementation, RNG consumption, tier priority independent of score scale, fit vs independent Cholesky). main_pilot3_frozen.go.txt: pilot 3 main.go for diff.
- train.json (pilot 2 artifact), weights_elig.json (output of ./he -mode fitelig -train train.json -weights weights_elig.json; reproducible).
- analyze4.py, run_pilot4.sh, PREREG.md, HASHES.txt (sha256 of every other file in the package; it does not hash itself).
- fixtures/: unit-test output, fit output, numpy cross-check script and output, load check (seed 951, 40,000 expansions, not evidence).
Reproduce checks: go vet ./... ; go test ./... ; ./he -mode fitelig ... ; python3 fixtures/numpy_crosscheck.py.

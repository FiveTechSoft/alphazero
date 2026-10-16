# Separate research variants, 2026-10-16

These branches are source preparations, not a merged upgrade. New variants have NOT been compiled or tested. Static source/embedded-asset inspection is not compilation. Remote validation remains pending under the existing bounded-resource protocol.

Base:7276f699e48276d5761a4e75b5f03a5fa9cb2aab. This is a new reimplementation, not recovery of an old unversioned package. Old results do not validate the new variants.

- research/historical-traince-preserved: original historical updater; full validation tests are retained outside the public repository.
- research/traince-joint-contract: joint gradient, one Adam transition per sample, fixed B1/B2.
- research/architecture24-utility-bidirectional: utility pruning, verified reversible backward search, hindsight and bestCost.
- fix/mcts-expanded-perspective: minimal Expanded flag fix, optional single-player value convention.
- fix/guidedsolve-best-cost: strict relaxation, reopening and stale heap entry rejection.
- fix/wasm-persisted-library: load the existing embedded lemma library instead of an empty library.

Do not silently combine branches. Architecture/search/TTT variants retain the historical updater. The optimizer comparison must use an unchanged solver on both arms.

Preserved SHA256:
- net.go H:0b4f2aee0b09431512d8305956c49320c97b0ceb6b87acefac884a7c0479576f
- traince_contract_test.go:6ec8fe3517cb28f191c7b98e4a4a5a58ed1b2e339588cfb3343298b12e7d1e14
- traince_observation_test.go:d5cbf0b49c0eba791e4909ced7b18fdf8cefc6065582b30dec48b14e547ce377

Historical replay belongs on H. It intentionally differs from C and must not be selected for C validation. Record actual commit/source hash rather than inferring branch identity from preserved test log labels. Resource-gated tests must not be weakened merely to obtain a local PASS.

Original DISENO_AUTOMEJORA.md and its measured historical claims remain intact. New notes here do not certify or rewrite those claims.

Public branches intentionally exclude operationally gated research tests. Public source trees are not byte-identical archives of the full preserved validation package. Test hashes above identify external validation sources, not files present here.

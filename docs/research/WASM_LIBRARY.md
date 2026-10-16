# Use persisted lemma library in WASM

Prepared source, NOT compiled or browser-tested. Replace empty NewLibrary startup with LoadLibrary("lemas.json"). Existing readDataFile fallback already embeds lemas.json; loader verifies insertions with unchanged canonEq and structural index. No fetched remote data, new dependency, rules or softmax change.

This changes the web solver's active library; it is a separate branch, not merged into historical/C/comparison trees. Existing fixed library increases actions and may affect branching/costs. Functional browser/regression and bounded performance validation remain pending. Does not claim better solving merely because library loads.

Test names/gates refer to the complete preserved validation suite, intentionally not included in this public source branch. No validation result is claimed.

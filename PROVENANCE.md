# mathzero-next, incremento 1 (P1 inicial)
Base: FiveTechSoft/alphazero rama research/architecture24-utility-bidirectional, commit 62c089df267f8862e2b9e10e44aeb008e6df79de
(commit declarado "unvalidated" por su autor). Rama destino propuesta: exp/mathzero-next, main intacto.
Base verificada: los 32 ficheros de mathzero/go coinciden con los SHA de blob de ese arbol; go build, go vet y go test ./... OK (Go 1.26.10, 46.0 s).
Este incremento ANADE dos ficheros nuevos sobre esa base, sin modificar ninguno existente:
- mathzero/go/d1_experiment_adapter.go: EvalBank/EvalConfig/SummarizeBank. Evaluacion determinista de un banco de casos con presupuesto y biblioteca explicitos; no entrena, no muta la biblioteca, restaura estado global.
- mathzero/go/d1_experiment_adapter_test.go: 5 tests (determinismo, igualdad con GuidedSolve, restauracion de globales y biblioteca intacta, biblioteca vacia = nil, cap 0).
Son codigo NUEVO. No reproducen ni validan el adaptador de la campana historica Reduced8; ninguna cifra historica acredita este arbol.
Siguen ausentes en todas las ramas del repo (se escribiran como codigo nuevo): runner D1, learning_probe_test, traince_contract_test, traince_observation_test, architecture24_test.
Con el incremento aplicado: go test ./... OK.

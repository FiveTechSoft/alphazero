# mathzero-next, incremento 7 (digest versionado, entrenamiento P3, diseno congelado)
Base: exp/mathzero-next HEAD 48c87b8. Ficheros:
- NUEVO mathzero/go/d1_train.go y d1_train_test.go: TrainRounds (autojuego bidireccional + TrainCE, sin biblioteca, sin ficheros, sin evaluacion ni seleccion de checkpoints; ignora y restaura LemmaLib/SearchTraining), WeightsSHA (pesos exportados + Level + Round; no estado del optimizador), ExcludeSetFor/RenderExcludeFor, OverlapTrainEval. 6 tests: determinismo y cambio de pesos, exclusion de pares de evaluacion, biblioteca ignorada y globales restaurados, EvalBank no cambia pesos, configuracion invalida, conteo de solapes de Start/Target/par.
- MODIFICADO mathzero/go/bank_gen.go: constante BankDigestAlgo.
- MODIFICADO mathzero/go/d1_runner_main.go: el informe anade bank_digest y bank_digest_algo (semantica unica: sha256 sobre lineas nivel\0inicio\0objetivo\n en orden, RenderPrefix, sin ids). Se CONSERVA bank_sha256 con su semantica historica: modo levels.json = sha256 de lineas inicio\0objetivo\n (sin nivel); modo generado = sha256 de la cadena hex de BankDigest (hash de un hash). No se recalculan resultados ya guardados (pilots/p2-library-budget conserva sus bank_sha256).
- MODIFICADO mathzero/go/d1_runner_main_test.go: prueba que bank_digest coincide con su algoritmo declarado recalculado de forma independiente, y que la semantica historica de bank_sha256 no cambia, en ambos modos.
- NUEVO pilots/p3/DESIGN-P3-frozen.md y P3-cost-measurements.md; scripts/guard.sh.
Verificado local (Go 1.26.10, bajo guard.sh con limite 900 MiB): gofmt y vet limpios, go test -v ./... ok 80.4 s, 43 PASS, 0 FAIL, 0 SKIP, RSS maximo 263 MiB, 0 procesos residuales.
Sin equivalencia con Reduced8; no hay resultados de aprendizaje en este incremento.

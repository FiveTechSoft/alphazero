# mathzero-next, incremento 6 (P2: generador de banco con hash y piloto de coste)
Base: exp/mathzero-next HEAD bd00677. Dos ficheros NUEVOS y una MODIFICACION de un fichero propio del incremento 2:
- NUEVO mathzero/go/bank_gen.go: GenBank(BankRule{Seed,PerLevel,Levels}) sobre genTrain de la base; regla declarada (un rand por semilla, niveles en orden, rechaza pares presentes en levels.json, blind.json o ya extraidos; identidad de par = Key normalizada AC de inicio y objetivo; sin casos identidad); BankDigest (SHA256 del contenido ordenado, no de los ids); BankOverlap (casos, inicios distintos, pares distintos, pares compartidos entre bancos, pares en bancos fijos).
- NUEVO mathzero/go/bank_gen_test.go: 4 tests (determinismo/composicion, exclusion de bancos fijos y duplicados, digest sobre contenido, conteo de solapes).
- MODIFICADO mathzero/go/d1_runner_main.go: flags -bankseed -perlevel -levels para usar GenBank; el informe anade bank_rule, nodes_total, elapsed_ms (elapsed_ms no es determinista).
Pilotos locales (CPU, <=4 GB): ver el informe al responsable; son mediciones de coste, sin pre/post ni evidencia de aprendizaje.
Sin equivalencia con Reduced8. Verificado local (Go 1.26.10): gofmt y vet limpios, go test -v ./... ok 69.3 s, 36 PASS, 0 FAIL, 0 SKIP.

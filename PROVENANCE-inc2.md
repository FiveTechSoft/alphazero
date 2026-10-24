# mathzero-next, incremento 2 (runner P1)
Base: exp/mathzero-next HEAD 5f4d533 (= 62c089df267f8862e2b9e10e44aeb008e6df79de + incremento 1).
Anade 2 ficheros nuevos, sin modificar ninguno existente (main.go intacto; el subcomando se engancha con init()):
- mathzero/go/d1_runner_main.go: `mathzero d1eval -level K -cap N [-model p | -seed S] [-lib p] [-max M]` imprime un JSON con configuracion, SHA256 del banco, del modelo y de la biblioteca (si se usan), resultados por caso y resumen. Usa EvalBank del incremento 1.
- mathzero/go/d1_runner_main_test.go: 4 tests (determinismo y coherencia del resumen, el digest del banco depende de la seleccion, entradas invalidas rechazadas, cableado del digest de biblioteca).
Codigo NUEVO. No reproduce ni valida el runner de la campana historica Reduced8; ninguna cifra historica acredita este arbol. Sin resultados de rendimiento.
Verificado local (Go 1.26.10): gofmt y vet limpios, go test ./... ok.

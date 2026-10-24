# mathzero-next, incremento 5 (observabilidad de coste, cierre P1)
Base: exp/mathzero-next HEAD 1160b74. Anade 1 fichero nuevo, sin modificar ninguno existente:
- mathzero/go/bank_observability_test.go:
  TestObservabilityNodesWithinCapAndSummary (invariante general): nodos <= cap en cada caso; SummarizeBank coherente (N, Solved, BudgetExhausted) en niveles 4 y 6, caps 20 y 40.
  TestObservabilityMeasuredBankProfile (MEDIDO, no teorema): con MLP semilla 11 y los 12 primeros casos de cada nivel, cap 20: nivel 4 = 11 resueltos / 1 agotado, nivel 6 = 2 / 10; todo caso no resuelto uso su presupuesto completo; subir cap de 20 a 40 no perdio ningun caso resuelto.
Medicion de apoyo (no fijada en tests; MLP semilla 11, 12 casos, 1 ejecucion): nivel 4 resueltos 11/12/12/12 con cap 20/40/80/160; nivel 6 resueltos 2/3/6/10 con cap 20/40/80/160, tiempos 1.6/3.8/6.0/12.1 s; 0 casos con nodos>cap y 0 fallos por debajo del cap; 0 violaciones de monotonia entre caps consecutivos en este banco.
Son medidas de este banco y esta red sin entrenar; no generalizan ni validan nada de Reduced8.
Verificado local (Go 1.26.10): gofmt y vet limpios, go test -v ./... ok 60.5 s, 32 PASS, 0 FAIL, 0 SKIP.

# mathzero-next, incremento 3 (tests de contrato)
Base: exp/mathzero-next HEAD 824f608. Anade 1 fichero nuevo, sin modificar ninguno existente:
- mathzero/go/bidir_contract_test.go: 9 tests de invariantes del codigo heredado de research/architecture24-utility-bidirectional (62c089d, "unvalidated" segun su autor).
  Evaluacion: nunca produce hindsight ni muestras en busqueda fallida; nodos <= cap; identidad = 0 nodos; cap 0 no resuelve; la evaluacion no toca contadores de biblioteca.
  Entrenamiento: indices de muestra validos; resuelto/no resuelto igual que en modo evaluacion (sin biblioteca, 16 casos de niveles 2 y 4).
  Aristas: exactAction recupera cada accion enumerada y verifiedEdge la acepta; backwardEdgeAllowed rechaza las reglas 2, 10, 12, 13.
  Utilidad: formula Utility = aplicaciones_ganadoras*(ahorro-1) - lambda*nodos_generados_fallidos; PruneUtility descarta utilidad<=0 con evidencia suficiente, conserva los de poca evidencia, ordena por utilidad, respeta ActiveCap y mantiene ByKey coherente.
Hallazgo documentado por el test: una busqueda RESUELTA devuelve las muestras de su camino en cualquier modo; solo el fallo (y el hindsight) son exclusivos de entrenamiento.
Son pruebas de invariantes, no de rendimiento. No reproducen los tests de la campana historica Reduced8; sin equivalencia historica.
Verificado local (Go 1.26.10): gofmt y vet limpios, go test ./... ok 48.2 s.

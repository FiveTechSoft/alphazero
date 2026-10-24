# Validacion P3, seeds 11..18 (RESULTADO PREREGISTRADO — n=8)

Salida cruda (depurada) de la validacion congelada 11..18 del diseno P3
(pilots/p3/DESIGN-P3-frozen.md), ejecutada por etapas bajo guardia inc8 + wrapper inc11,
tras el desbloqueo posterior al piloto 101. Piloto 101 excluido del contraste, segun
preregistro.

Resultado medido (primario, 90 casos por semilla):
  seed 11: 53 -> 59 (+9/-3)   seed 15: 42 -> 67 (+25/-0)
  seed 12: 46 -> 63 (+17/-0)  seed 16: 48 -> 61 (+13/-0)
  seed 13: 43 -> 64 (+21/-0)  seed 17: 41 -> 53 (+12/-0)
  seed 14: 51 -> 64 (+14/-1)  seed 18: 39 -> 49 (+12/-2)

Criterio preregistrado: 8/8 deltas primarios positivos. Test de signos bilateral
nominal p = 0.0078125, condicional a bancos fijos. Secundario (descriptivo):
deltas +12, +16, +17, +16, +23, +14, +9, +4.

Auditoria independiente realizada antes de esta persistencia:
- Integridad: 48/48 etapas COMPLETED, 0 censuras, guardias sin procesos residuales,
  presupuestos 189-247 s de 900 s, binario fijo 60a95de7 en todas las ejecuciones.
- Recomputo desde los datos: digests de ambos bancos (1bd3a2da.../8909948a...,
  diseno congelado), config (r6x40, cap 100/100, lib=none), pesos pre (NewMLP(seed))
  y post (LoadMLP de weights_post.json), aritmetica primaria y secundaria.
- Exclusiones recomputadas: PairInEval = 0 en las 16 combinaciones semilla x banco
  (requisito duro del protocolo, cumplido).
- Rejuicio acotado: 82 casos de transicion (las 8 perdidas + 74 ganancias muestreadas
  1 de cada 3), ambas redes, igualdad exacta de CaseResult incluidos nodos.
- EvalBank no muta pesos (hash pre/post de evaluacion identico).

LIMITACIONES:
- n=8: validacion pequena; el contraste es solo el test de signos preregistrado,
  condicional a bancos fijos. Sin control SGD; cap 100 igual en train/eval.
- Solapes Start (transferencia cercana, no generalizacion limpia) — recomputados,
  coinciden con lo extraido por el ejecutor:
    primario   StartInEval: 86, 97, 76, 84, 75, 77, 86, 76 (Target 0 en todas)
    secundario StartInEval: 86, 106, 86, 100, 87, 89, 91, 81
  Esto NO viola el protocolo congelado: el diseno excluye PARES exactos (ambas
  orientaciones) y explicitamente no garantiza separacion de Start ni de Target
  (DESIGN-P3-frozen.md); los solapes se registran por semilla, como aqui.
- Excepcion Target en secundario: seeds 12 y 15 tienen TargetInEval = 1 (0 en el
  resto, primario y secundario). Documentado; permitido por el protocolo.
- Bancos no independientes: 118 starts distintos entre 180 casos; 30 starts del
  secundario aparecen tambien en el primario (por diseno).
- Niveles de entrenamiento 2..7 se solapan con evaluacion 6..8.
- NO es demostracion de aprendizaje en general ni equivalencia con ningun resultado
  historico (Reduced8 o anterior).
- Binario no reconstruido byte a byte en la auditoria (toolchain distinta); la
  reproduccion funcional (solved y nodos exactos) y el hash del binario son
  evidencias distintas, ninguna sustituye a la otra.

Transformacion respecto al paquete recibido: HASHES.txt depurado (rutas de trabajo
eliminadas, hashes conservados); a los 5 stage JSON de cada semilla se anadio un
newline final (el editor web lo impone; contenido JSON intacto). Los hashes de
report.out y weights_post.json en HASHES.txt corresponden a los archivos aqui
presentes (sin cambios byte a byte). MANIFEST.sha256 refleja los bytes aqui
presentes y no se lista a si mismo; su hash se registra por separado fuera del archivo.

# Hito 14 - Benchmark de lenguaje (bucle caliente)

> **CORRECCION (26 oct 2026): la cifra de Go de la tabla original era invalida.**
> `expand()` no marcaba `Node.Expanded = true`, asi que el MCTS de Go no
> exploraba: devolvia la primera accion legal y el "bench" medía un bucle
> vacio. C y Rust sí exploraban. Corregido en el commit `36d01d2` (test
> `TestMCTSExpandeRaizYVisita`). Re-medido en esta maquina (Go 1.27.3,
> gcc -O2, rustc 1.99.1), mediana de 5 ejecuciones:
>
> | Lenguaje | sims/s | partidas/s | vs Go |
> |----------|--------|------------|-------|
> | Go 1.27.3 | 1.278.000 | ~1.826 | 1.00x |
> | C (-O2) | 3.700.000 | ~5.288 | 2.90x |
> | Rust (-O) | 4.195.000 | ~6.028 | 3.28x |
>
> En esta maquina Rust queda ~13% por encima de C y ambos ~3x Go. Las
> cifras absolutas no son comparables con las del 12-oct (otra maquina).
> Lectura: el fix no cambia el veredicto cualitativo previo (C/Rust
> claramente mas rapidos en el bucle caliente), pero invalida los
> multiplicadores publicados antes.


Medido el 12 oct 2026, misma maquina (sandbox), mismo algoritmo y
parametros en los tres lenguajes: MCTS PUCT sobre tres en raya sin red
(prior uniforme), 200 partidas x 100 simulaciones por jugada,
c_puct=1.5, tie-break por primer maximo. Los tres recorren la MISMA
trayectoria (total identico: 140.000 simulaciones), asi que la
comparacion es del bucle, no de la logica.

Comandos repetibles:
- Go:   `cd mathzero/go && go build -o mathzero . && ./mathzero benchttt`
- C:    `gcc -O2 -o bench_ttt_c bench/bench_ttt.c -lm && ./bench_ttt_c`
- Rust: `rustc -O -o bench_ttt_rs bench/bench_ttt.rs && ./bench_ttt_rs`

Resultado ORIGINAL del 12-oct (INVALIDO para Go, ver correccion arriba; mediana de 5 ejecuciones):

| Lenguaje | sims/s      | partidas/s | vs Go |
|----------|-------------|------------|-------|
| Go 1.26  | 2.928.724   | ~4.100     | 1.00x |
| C (-O2)  | 5.512.827   | ~7.400     | 1.88x |
| Rust (-O)| 5.742.652   | ~8.300     | 1.96x |

Lectura honesta:
- En el bucle caliente puro, Rust y C corren ~2x que Go. Rust ~4% por
  encima de C (dentro del ruido: en ejecuciones sueltas se cruzan).
- El port de Go usa map[int]*Node para los hijos (lo idiomatico en el
  motor actual); C y Rust usan arenas de arrays. Parte de la diferencia
  es de estructura de datos, no solo de lenguaje.
- Con red conectada, la prediccion domina el tiempo y la diferencia de
  lenguaje se estrecha: esto mide el limite favorable a C/Rust.
- NO medido: el juez de MathRing (canon de polinomios) ni la busqueda
  guiada; portarlos es un trabajo mayor (arboles de expresiones,
  hashing de claves). El MCTS es el bucle dominante en todas las
  busquedas actuales.

Conclusion (revisada tras la correccion): en el bucle caliente C y Rust
dan ~3x Go; Rust queda algo por encima de C y aporta seguridad de memoria.
La decision de lenguaje sigue abierta y debe re-medirse en la maquina de
decision, con el mismo binario y sin el bug de `Expanded`.

# Hito 14 - Benchmark de lenguaje (bucle caliente)

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

Resultado (mediana de 5 ejecuciones):

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

Conclusion: si el proyecto migra, Rust da ~2x en el bucle caliente con
seguridad de memoria; C da lo mismo sin ella. La decision queda abierta
y ahora tiene numeros.

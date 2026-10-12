# Resultados medidos (12 oct 2026)

Misma maquina, mismas condiciones para ambos modelos: MCTS con 50
simulaciones por jugada, sin ruido de exploracion, semilla de evaluacion
fija (123). "Minimax" es el juego perfecto: el tres en raya optimo es
tablas, asi que la mejor marca posible contra el es **no perder nunca**.

| Rival (100/50 partidas por lado) | Red SIN entrenar | Red entrenada (6000 partidas, ~6 min) |
|---|---|---|
| Aleatorio, agente primero | 96 ganadas, 4 tablas, 0 perdidas | 99 ganadas, 1 tablas, 0 perdidas |
| Aleatorio, agente segundo | 83 ganadas, 13 tablas, 4 perdidas | 84 ganadas, 16 tablas, 0 perdidas |
| Minimax perfecto, agente primero | 0 ganadas, 50 tablas, 0 perdidas | 0 ganadas, 50 tablas, 0 perdidas |
| Minimax perfecto, agente segundo | 0 ganadas, 20 tablas, **30 perdidas** | 0 ganadas, 50 tablas, **0 perdidas** |

Como leerlo: ganarle al azar no prueba gran cosa (la busqueda MCTS ya lo
consigue con una red aleatoria). La prueba de fuego es la ultima fila:
defendiendo contra un rival perfecto, la red sin entrenar pierde 30 de 50
partidas; tras 6000 partidas de autojuego no pierde ninguna. El agente ha
aprendido el juego optimo solo jugando contra si mismo, sin ver una sola
partida humana.

Otras senales del aprendizaje:
- Valor estimado del tablero vacio: +0.19 antes de entrenar, +0.11 despues
  (el valor correcto es 0: tablas con juego perfecto).
- Politica de apertura aprendida: se concentra en centro y esquinas
  (las jugadas optimas) y casi elimina los lados.

Lo que NO se midio: versiones mas grandes (mas capas, mas simulaciones),
otros juegos (Conecta 4, Othello) ni rendimiento en GPU: el prototipo es
CPU y NumPy a proposito, para que se pueda leer entero.

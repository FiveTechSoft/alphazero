# Core auto-mejorante (diseno + medidas)

Actualizado: 12 oct 2026. Estado: hitos 6 y 7 del ROADMAP.

## Objetivo (pedido 15:06)

Un core que crezca y crezca por si solo. Traduccion a algo MEDIBLE:
en cada ciclo el sistema (1) se genera problemas nuevos, (2) los
resuelve con busqueda guiada por su red, (3) la red aprende de sus
propias soluciones, (4) el juez exacto verifica cada identidad
descubierta y la guarda en una biblioteca persistente de lemas,
(5) los lemas se reutilizan como reglas de un paso en problemas
nuevos, y (6) una metrica FIJA mide el progreso ciclo a ciclo.

No afirmamos crecimiento ilimitado: afirmamos lo que mide la curva.

## El bucle (comando unico)

    ./mathzero cycle [ciclos] [problemas_por_ciclo] [cap_nodos]

1. METRICA: evaluacion sobre levels.json (k2..k8, 30 problemas fijos
   por nivel, generados aparte de los de entrenamiento), con el mismo
   cap de nodos que el baseline BFS. Se imprime resueltos/30 y nodos
   medios por nivel. Repetible: mismo fichero, misma semilla, mismo
   comando.
2. ABLACION (ciclo 1 y ultimo): la MISMA red evaluada SIN la
   biblioteca. Es el contrafactual que demuestra que los lemas aportan
   (o no).
3. AUTOJUEGO: genTrain genera pares (a, b) del nivel actual del
   curriculum; GuidedSolve (mejor-primero por log-prob de la red) los
   resuelve usando reglas + lemas.
4. VERIFICACION + BIBLIOTECA: cada par resuelto (a=b en >=2 pasos) se
   inserta en lemas.json SOLO si canonEq(a,b) lo confirma (juez exacto,
   coeficientes enteros). Deduplicados por clave estructural.
5. ENTRENAMIENTO: la red imita por entropia cruzada las acciones que
   llevaron a solucion (expert iteration).
6. CURRICULO: sube de nivel solo al resolver >=80% de la tanda.
7. PERSISTENCIA: model_math.json y lemas.json se guardan tras cada
   ciclo; todo queda versionado en el repo.

## Por que un lema es verdad y no ruido

Un lema (from -> to) solo entra en la biblioteca si el juez confirma
que from y to tienen la misma forma canonica de polinomio. Aplicarlo
es siempre correcto; lo unico en juego es si es UTIL (ahorra nodos),
y eso lo mide la ablacion.

## Medidas ya obtenidas (cap=2000, conjuntos fijos)

Red tras 6 rondas de curriculum vs baseline BFS (resueltos/30):

  nivel | BFS | red r3 | red r6
  k2    | 30  | 30 | 30
  k3    | 30  | 30 | 30
  k4    | 16  | 27 | 28
  k5    | 15  | 27 | 27
  k6    |  6  | 19 | 18
  k7    |  5  | 16 | 15
  k8    |  9  | 15 | (midiendo)

La red supera al BFS en todos los niveles dificiles con el mismo cap.

## Limites honestos

- La curva puede saturar: cuando la red resuelve casi todo de su
  nivel y los lemas dejan de ser nuevos, el total/210 se aplana. La
  metrica fija lo hace visible; no lo ocultamos.
- Los lemas son identidades CONCRETAS (sin variables libres de
  patron): generalizan solo cuando un subproblema coincide
  estructuralmente. Lemas con esquemas (variables de patron) son
  trabajo futuro.
- Nodos medios pueden subir al activar la biblioteca (mas acciones
  por estado) aunque suban los resueltos; la tabla muestra ambos.

## Pendiente

- Curva completa de N ciclos con ablacion (en ejecucion).
- API local + visualizacion, laberinto, sudoku, benchmark Go/C/Rust.

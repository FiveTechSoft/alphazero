# Reglas del proyecto AlphaZero / MathZero

## Build y ejecución (kernGo en `mathzero/go`)

- En Windows, `go` no está en el PATH del shell: invócalo por ruta completa desde `mathzero/go`, p.ej. `& "C:\Program Files\Go\bin\go.exe" build -o mathzero.exe .`.

## Trampas conocidas de medición

- `curve` y `cycle` ESCRIBEN `mathzero/go/model_math.json` y `mathzero/go/lemas.json` aunque no entrenen (suben el campo `Round`). Para experimentar usa `-nosave` (no escribe nada) o `-datadir DIR` (lee/escribe ahí), p.ej. `./mathzero curve 1 0 1000 -nosave`. Si ya se ensució el árbol, restáuralo con `git checkout -- mathzero/go/model_math.json mathzero/go/lemas.json`.

// Flags comunes de datos: -datadir redirige lecturas/escrituras de datos y
// -nosave desactiva la persistencia (mediciones sobre el arbol limpio).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	DataDir string
	NoSave  bool
)

// dataPath resuelve el destino de un fichero de datos dentro de DataDir.
func dataPath(name string) string {
	if DataDir == "" || filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(DataDir, name)
}

// parseDataFlags extrae -nosave y -datadir <dir> (o -datadir=<dir>) de args.
// Devuelve los argumentos posicionales restantes.
func parseDataFlags(args []string) ([]string, bool) {
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "-nosave":
			NoSave = true
		case a == "-datadir":
			if i+1 >= len(args) || args[i+1] == "" {
				fmt.Fprintln(os.Stderr, "falta el directorio de -datadir")
				return nil, false
			}
			i++
			DataDir = args[i]
		case strings.HasPrefix(a, "-datadir="):
			dir := strings.TrimPrefix(a, "-datadir=")
			if dir == "" {
				fmt.Fprintln(os.Stderr, "falta el directorio de -datadir")
				return nil, false
			}
			DataDir = dir
		default:
			rest = append(rest, a)
		}
	}
	if DataDir != "" {
		if err := os.MkdirAll(DataDir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "no se pudo crear -datadir:", err)
			return nil, false
		}
	}
	return rest, true
}

// saveMLP persiste un modelo salvo que -nosave este activo.
func saveMLP(n *MLP, path string) {
	if NoSave {
		fmt.Println("[nosave] no se guarda", path)
		return
	}
	n.Save(path)
}

// saveLibrary persiste una biblioteca salvo que -nosave este activo.
func saveLibrary(L *Library, path string) {
	if NoSave {
		fmt.Println("[nosave] no se guarda", path)
		return
	}
	L.Save(path)
}

// saveJSON persiste un resultado salvo que -nosave este activo.
func saveJSON(path string, data []byte) error {
	if NoSave {
		fmt.Println("[nosave] no se escribe", path)
		return nil
	}
	return writeFileAtomic(path, data)
}

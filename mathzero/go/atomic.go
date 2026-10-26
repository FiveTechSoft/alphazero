package main

import (
	"os"
	"path/filepath"
)

// writeFileAtomic escribe en un temporal del mismo directorio y hace
// rename: un fallo a mitad de escritura no deja el modelo truncado.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Chmod(0o644); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		// En algunos sistemas (p.ej. Windows) Rename puede no reemplazar
		// un fichero existente; se conserva el contenido anterior si falla.
		if _, statErr := os.Stat(path); statErr != nil {
			return err
		}
		if rmErr := os.Remove(path); rmErr != nil {
			return err
		}
		return os.Rename(tmp, path)
	}
	return nil
}

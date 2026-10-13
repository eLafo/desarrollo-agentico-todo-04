package actions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"

	"todo/public"
)

// assetVersionLength es el número de caracteres hexadecimales del SHA-256 que
// forman la versión de un asset.
const assetVersionLength = 12

// assetVersion devuelve la versión del fichero name de fsys: los primeros
// caracteres hexadecimales del SHA-256 de su contenido. El mismo contenido da
// siempre la misma versión y un contenido distinto da otra.
func assetVersion(fsys fs.FS, name string) (string, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:assetVersionLength], nil
}

// assetURL es el helper de plantilla que enlaza un fichero de public/assets con
// la versión de su contenido, por ejemplo /assets/app.css?v=51e6b6361204. Lee
// el mismo sistema de ficheros que sirve app.ServeFiles y lo hace en cada
// render, así que la URL cambia en cuanto cambia lo que se sirve y el navegador
// no reutiliza la copia en caché de otro contenido.
func assetURL(name string) (string, error) {
	file := path.Join("assets", name)
	version, err := assetVersion(public.FS(), file)
	if err != nil {
		return "", fmt.Errorf("no se puede versionar el asset %s: %w", name, err)
	}
	return "/" + file + "?v=" + version, nil
}

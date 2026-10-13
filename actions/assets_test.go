package actions

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"regexp"
	"testing/fstest"

	"todo/public"
)

// Enlaces que debe generar el layout: la versión son 12 caracteres hexadecimales.
var (
	stylesheetLink = regexp.MustCompile(`<link rel="stylesheet" href="(/assets/app\.css\?v=([0-9a-f]{12}))">`)
	scriptTag      = regexp.MustCompile(`<script src="(/assets/app\.js\?v=([0-9a-f]{12}))" defer></script>`)
)

// contentVersion calcula, sin pasar por el helper, la versión que corresponde a
// un contenido: los 12 primeros caracteres hexadecimales de su SHA-256.
func contentVersion(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:12]
}

func (as *ActionSuite) Test_AssetVersion_DerivesFromContent() {
	// Dos "ramas" con el mismo nombre de fichero y distinto CSS, y una copia exacta de la primera.
	branchA := fstest.MapFS{"assets/app.css": {Data: []byte(".nav-link { background: #f7e9e4; }\n")}}
	branchB := fstest.MapFS{"assets/app.css": {Data: []byte(".nav-link[aria-current=\"page\"] { background: #f7e9e4; }\n")}}
	copyOfA := fstest.MapFS{"assets/app.css": {Data: []byte(".nav-link { background: #f7e9e4; }\n")}}

	a, err := assetVersion(branchA, "assets/app.css")
	as.NoError(err)
	as.Regexp(`^[0-9a-f]{12}$`, a)
	as.Equal(contentVersion(branchA["assets/app.css"].Data), a)

	b, err := assetVersion(branchB, "assets/app.css")
	as.NoError(err)
	as.NotEqual(a, b) // otro contenido con el mismo nombre: otra URL

	same, err := assetVersion(copyOfA, "assets/app.css")
	as.NoError(err)
	as.Equal(a, same) // mismo contenido: misma URL, la caché sigue valiendo

	_, err = assetVersion(branchA, "assets/no-existe.css")
	as.ErrorIs(err, fs.ErrNotExist)
}

func (as *ActionSuite) Test_AssetURL_VersionsServedFile() {
	for _, name := range []string{"app.css", "app.js"} {
		data, err := fs.ReadFile(public.FS(), "assets/"+name)
		as.NoError(err)

		got, err := assetURL(name)
		as.NoError(err)
		as.Equal("/assets/"+name+"?v="+contentVersion(data), got)
	}

	_, err := assetURL("no-existe.css")
	as.ErrorIs(err, fs.ErrNotExist)
}

func (as *ActionSuite) Test_Layout_LinksAssetsByContentVersion() {
	task := as.taskForEditing(false)

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()

	// Ya no hay versiones elegidas a mano. La comilla de cierre evita falsos
	// positivos: una huella puede empezar por 4 o por 2.
	as.NotContains(body, `href="/assets/app.css?v=4"`)
	as.NotContains(body, `src="/assets/app.js?v=2"`)

	css := stylesheetLink.FindStringSubmatch(body)
	as.Len(css, 3, "el layout debe enlazar app.css con la versión de su contenido")
	js := scriptTag.FindStringSubmatch(body)
	as.Len(js, 3, "el layout debe enlazar app.js con la versión de su contenido")

	// Las tres pantallas de las capturas enlazan las mismas URLs.
	for _, page := range []string{"/eisenhower", "/tasks/" + task.ID.String() + "/edit"} {
		pageBody := as.HTML("%s", page).Get().Body.String()
		as.Contains(pageBody, css[0], page)
		as.Contains(pageBody, js[0], page)
	}

	// La versión de cada URL es la huella exacta de lo que el servidor sirve en ella...
	cssRes := as.HTML("%s", css[1]).Get()
	as.Equal(200, cssRes.Code)
	as.Equal(css[2], contentVersion(cssRes.Body.Bytes()))
	jsRes := as.HTML("%s", js[1]).Get()
	as.Equal(200, jsRes.Code)
	as.Equal(js[2], contentVersion(jsRes.Body.Bytes()))

	// ...y lo que se sirve es la hoja de esta rama, con los estilos de la propuesta.
	sheet := cssRes.Body.String()
	as.Contains(sheet, ".eisenhower-grid")
	as.Contains(sheet, ".task-dimensions")
	as.Contains(sheet, `.nav-link[aria-current="page"]`)
}

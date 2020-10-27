package views

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"path/filepath"

	"github.com/iamtraining/gallery/context"
)

var (
	LayoutDir string = "views/layouts/"
	TmplDir   string = "views/"
	TmplExt   string = ".gohtml"
)

type View struct {
	Tmpl   *template.Template
	Layout string
}

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, extract()...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Tmpl:   t,
		Layout: layout,
	}
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	var d Data

	switch ok := data.(type) {
	case Data:
		d = ok
	default:
		data = Data{
			Body: data,
		}
	}

	d.User = context.GetUser(r.Context())

	var buf bytes.Buffer

	err := v.Tmpl.ExecuteTemplate(&buf, v.Layout, data)
	if err != nil {
		http.Error(w, "something goes wrong", http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}

func extract() []string {
	tmpls, err := filepath.Glob(LayoutDir + "*" + TmplExt)
	if err != nil {
		panic(err)
	}

	return tmpls
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TmplDir + f
	}
}

func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TmplExt
	}
}

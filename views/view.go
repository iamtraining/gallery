package views

import (
	"html/template"
	"net/http"
	"path/filepath"
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

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Tmpl.ExecuteTemplate(w, v.Layout, data)
}

func extract() []string {
	tmpls, err := filepath.Glob(LayoutDir + "*" + TmplExt)
	if err != nil {
		panic(err)
	}

	return tmpls
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.Render(w, nil); err != nil {
		panic(err)
	}
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

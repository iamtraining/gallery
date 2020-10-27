package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/iamtraining/gallery/context"
	"github.com/iamtraining/gallery/models"
	"github.com/iamtraining/gallery/views"
)

const (
	ShowGallery  = "show_gallery"
	IndexGallery = "index_gallery"
	EditGallery  = "edit_gallery"
	Multipart    = 1 << 20
)

type Galleries struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	g         models.GalleryService
	r         *mux.Router
	i         models.ImgService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

func NewGalleries(g models.GalleryService, i models.ImgService, r *mux.Router) *Galleries {
	return &Galleries{
		New:       views.NewView("bootstrap", "galleries/new"),
		ShowView:  views.NewView("bootstrap", "galleries/show"),
		EditView:  views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		g:         g,
		r:         r,
		i:         i,
	}
}

// POST /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data views.Data

	var form GalleryForm

	if err := parseForm(r, &form); err != nil {
		data.SetAlert(err)
		g.New.Render(w, r, data)
		return
	}

	user := context.GetUser(r.Context())

	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}

	if err := g.g.Create(&gallery); err != nil {
		data.SetAlert(err)
		g.New.Render(w, r, data)
		return
	}

	url, err := g.r.Get(ShowGallery).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid gallery ID", http.StatusNotFound)
		return nil, err
	}

	gallery, err := g.g.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "something goes wrong", http.StatusInternalServerError)
		}
		return nil, err
	}

	img, _ := g.i.ByGalleryID(gallery.ID)
	gallery.Img = img

	return gallery, nil
}

func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	var data views.Data
	data.Body = gallery

	g.ShowView.Render(w, r, data)
}

func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.GetUser(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "you dont gave permission to edit this gallery", http.StatusForbidden)
		return
	}

	var data views.Data
	data.Body = gallery
	g.EditView.Render(w, r, data)
}

func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.GetUser(r.Context())

	if gallery.UserID != user.ID {
		http.Error(w, "gallery not found", http.StatusNotFound)
		return
	}

	//
	var data views.Data
	data.Body = gallery

	var form GalleryForm

	if err = parseForm(r, &form); err != nil {
		data.SetAlert(err)
		g.EditView.Render(w, r, data)
		return
	}

	gallery.Title = form.Title

	err = g.g.Update(gallery)
	if err != nil {
		data.SetAlert(err)
	} else {
		data.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "gallery successfully updated",
		}
	}

	g.EditView.Render(w, r, data)
}

func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.GetUser(r.Context())

	if gallery.UserID != user.ID {
		http.Error(w, "you dont have permissiont to delete this gallery", http.StatusForbidden)
		return
	}

	var data views.Data

	err = g.g.Delete(gallery.ID)
	if err != nil {
		data.SetAlert(err)
		data.Body = gallery
		g.EditView.Render(w, r, data)
		return
	}

	url, err := g.r.Get(IndexGallery).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	user := context.GetUser(r.Context())

	galleries, err := g.g.ByUserID(user.ID)
	if err != nil {
		http.Error(w, "something goes wrong", http.StatusInternalServerError)
		return
	}

	var data views.Data
	data.Body = galleries
	g.IndexView.Render(w, r, data)
}

// POST /galleries/:id/images -- jpg jpeg png
func (g *Galleries) UploadImg(w http.ResponseWriter, r *http.Request) {
	galelry, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.GetUser(r.Context())

	if galelry.UserID != user.ID {
		http.Error(w, "gallery not found", http.StatusNotFound)
		return
	}

	var data views.Data
	data.Body = galelry

	err = r.ParseMultipartForm(Multipart)
	if err != nil {
		data.SetAlert(err)
		g.EditView.Render(w, r, data)
		return
	}

	files := r.MultipartForm.File["files"]

	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			data.SetAlert(err)
			g.EditView.Render(w, r, data)
			return
		}

		defer file.Close()

		err = g.i.Create(galelry.ID, file, f.Filename)
		if err != nil {
			data.SetAlert(err)
			g.EditView.Render(w, r, data)
		}
	}

	data.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "images successfully uploaded",
	}

	g.EditView.Render(w, r, data)
}

func (g *Galleries) ImgDelete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.GetUser(r.Context())

	if gallery.UserID != user.ID {
		http.Error(w, "you dont have permission to delete this gallery", http.StatusForbidden)
		return
	}

	fname := mux.Vars(r)["filename"]

	i := models.Img{
		Filename:  fname,
		GalleryID: gallery.ID,
	}

	err = g.i.Delete(&i)
	if err != nil {
		var data views.Data
		data.Body = gallery
		data.SetAlert(err)
		g.EditView.Render(w, r, data)
		return
	}

	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

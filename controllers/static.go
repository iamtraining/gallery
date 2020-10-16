package controllers

import "github.com/iamtraining/gallery/views"

type Static struct {
	Home    *views.View
	Contact *views.View
}

func NewStatic() *Static {
	return &Static{
		Home: views.NewView(
			"bootstrap", "static/home"),
		Contact: views.NewView(
			"bootstrap", "static/contact"),
	}
}

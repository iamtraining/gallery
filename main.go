package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/iamtraining/gallery/controllers"
	"github.com/iamtraining/gallery/middleware"
	"github.com/iamtraining/gallery/models"
)

//601-627
//pagination?
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1111"
	dbname   = "gallery_dev"
)

func main() {
	psql := fmt.Sprintf("host=%s port =%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	serv, err := models.NewServices(psql)
	if err != nil {
		panic(err)
	}
	defer serv.Close()
	serv.DestructiveReset()

	r := mux.NewRouter()

	static := controllers.NewStatic()
	uc := controllers.NewUsers(serv.User)
	gc := controllers.NewGalleries(serv.Gallery, serv.Img, r)

	userMw1 := middleware.RequireUser{}

	userMw2 := middleware.User{
		UserService: serv.User,
	}

	newGallery := userMw1.Apply(gc.New)
	createGallery := userMw1.ApplyFn(gc.Create)

	r.NotFoundHandler = notfound()
	r.HandleFunc("/", static.Home.ServeHTTP).Methods("GET")
	r.HandleFunc("/contact", static.Contact.ServeHTTP).Methods("GET")
	r.HandleFunc("/faq", faq).Methods("GET")

	// user
	r.HandleFunc("/signup", uc.New).Methods("GET")
	r.HandleFunc("/signup", uc.Create).Methods("POST")
	r.Handle("/login", uc.LoginView).Methods("GET")
	r.HandleFunc("/login", uc.Login).Methods("POST")
	r.HandleFunc("/cookietest", uc.CookieTest).Methods("GET")

	// gallery
	r.Handle("/galleries/new", newGallery).Methods("GET")
	r.HandleFunc("/galleries", createGallery).Methods("POST").Name(controllers.IndexGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}", gc.ShowView.ServeHTTP).
		Methods("GET").Name(controllers.ShowGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", userMw1.ApplyFn(gc.Edit)).Methods("GET").Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", userMw1.ApplyFn(gc.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", userMw1.ApplyFn(gc.Delete)).Methods("POST")
	r.Handle("/galleries", userMw1.ApplyFn(gc.Index)).Methods("GET")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", userMw1.ApplyFn(gc.UploadImg)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", userMw1.ApplyFn(gc.ImgDelete)).Methods("POST")

	// fileserver
	imgHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images").Handler(http.StripPrefix("/images", imgHandler))

	http.ListenAndServe(":3000", userMw2.Apply(r))
}

func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "go to \"/\" to get to main page "+
		"go to \"/contact\" to get in touch with us")
}

func notfound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<h1>We could not find the page you "+
			"were looking for :(</h1>"+
			"<p>Please email us if you keep being sent to an "+
			"invalid page.</p>")
	}
}

package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/iamtraining/gallery/controllers"
	"github.com/iamtraining/gallery/models"
)

//326-352

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

	us, err := models.NewUserService(psql)
	if err != nil {
		panic(err)
	}
	defer us.Close()
	us.DestructiveReset()

	static := controllers.NewStatic()
	uc := controllers.NewUsers(us)

	r := mux.NewRouter()
	r.NotFoundHandler = notfound()
	r.HandleFunc("/", static.Home.ServeHTTP).Methods("GET")
	r.HandleFunc("/contact", static.Contact.ServeHTTP).Methods("GET")
	r.HandleFunc("/faq", faq).Methods("GET")
	r.HandleFunc("/signup", uc.New).Methods("GET")
	r.HandleFunc("/signup", uc.Create).Methods("POST")
	r.Handle("/login", uc.LoginView).Methods("GET")
	r.HandleFunc("/login", uc.Login).Methods("POST")
	r.HandleFunc("/cookietest", uc.CookieTest).Methods("GET")

	http.ListenAndServe(":3000", r)
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

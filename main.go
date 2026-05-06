package main

import (
	"log"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	// Render the home html page from static folder
	http.ServeFile(w, r, "static/home.html")
}

func coursePage(w http.ResponseWriter, r *http.Request) {
	// Render the course html page
	http.ServeFile(w, r, "static/courses.html")
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	// Render the about html page
	http.ServeFile(w, r, "static/about.html")
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	// Render the contact html page
	http.ServeFile(w, r, "static/contact.html")
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("static"))

	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/home", homePage)
	mux.HandleFunc("/courses", coursePage)
	mux.HandleFunc("/about", aboutPage)
	mux.HandleFunc("/contact", contactPage)

	return mux
}

func main() {
	err := http.ListenAndServe("0.0.0.0:8080", newMux())
	if err != nil {
		log.Fatal(err)
	}
}

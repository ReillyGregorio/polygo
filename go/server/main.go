package main

import (
	"flag"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// flags
var (
	port         = flag.String("port", ":8000", "HTTP service address (e.g., ':8000')")
	resourcesDir = flag.String("resources_dir", "", "The directory to find templates, JS, and CSS files. If blank the current directory will be used.")
	local        = flag.Bool("local", true, "Running locally, as opposed to in production.")
)

var (
	templates *template.Template
)

func makeResourceHandler() func(http.ResponseWriter, *http.Request) {
	fileServer := http.FileServer(http.Dir(*resourcesDir))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "max-age=300")
		fileServer.ServeHTTP(w, r)
	}
}

func loadTemplates() {
	templates = template.Must(template.New("").ParseFiles(
		filepath.Join(*resourcesDir, "templates/index.html"),
	))
}

type IndexData struct {
	Name string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infof("index.html")
	if *local {
		loadTemplates()
	}
	data := IndexData{
		Name: "Skeleton App",
	}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		glog.Errorf("Failed to expand template: %s", err)
	}
}

func main() {
	flag.Parse()

	// Resources are served directly.
	router := mux.NewRouter()
	router.PathPrefix("/res/").HandlerFunc(makeResourceHandler())

	// Add page handlers here.
	router.HandleFunc("/", indexHandler)

	http.Handle("/", router)
	glog.Infof("Server is running at: http://localhost%s", *port)
	glog.Fatal(http.ListenAndServe(*port, nil))
}

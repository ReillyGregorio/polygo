package main

import (
	"encoding/json"
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

type period struct {
	/*{period:"1st",class:"Math",classroom:"Room 511"},
	{period:"2nd",class:"CSP",classroom:"Room 2703"},
	{period:"3rd",class:"English",classroom:"Room 2305"},
	{period:"4th",class:"Civics",classroom:"Room 1711"},
	{period:"5th",class:"Homeroom",classroom:"Room 503"}*/
	Period    string `json:"period"`
	Class     string `json:"class"`
	Classroom string `json:"classroom"`
}

func classesHandler(w http.ResponseWriter, r *http.Request) {
	data := []period{
		period{
			Period:    "1st",
			Class:     "Math",
			Classroom: "Room 511",
		},
		period{
			Period:    "2nd",
			Class:     "CSP",
			Classroom: "Room 2703",
		},
		period{
			Period:    "3rd",
			Class:     "English",
			Classroom: "Room 2305",
		},
		period{
			Period:    "4th",
			Class:     "Civics",
			Classroom: "Room 1711",
		},
		period{
			Period:    "5th",
			Class:     "Homeroom",
			Classroom: "Room 503",
		},
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		glog.Errorf("Failed to encode: %s", err)
	}
}

type messages struct {
	/*
			{name:"Reilly:",msg:"What was the homework yesterday"},
		  	{name:"Isiah:",msg:"pages 5-17 I think"},
		  	{name:"Reilly:",msg:"oh right thx"},
		  	{name:"Generic Freshman:",msg:"I like cheese"},
	*/
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	data := []messages{
		messages{
			Name: "Reilly:",
			Msg:  "What was the homework yesterday",
		},
		messages{
			Name: "Isiah:",
			Msg:  "pages 5-17 I think",
		},
		messages{
			Name: "Reilly:",
			Msg:  "oh right thx",
		},
		messages{
			Name: "Generic Freshman::",
			Msg:  "WI like cheese",
		},
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		glog.Errorf("Failed to encode: %s", err)
	}
}

type events struct {
	/*
			{dow:"S",date:"Dec 24th 2017", hw:"Sleep"},
		  	{dow:"M",date:"Dec 25th 2017", hw:"Chapter 8 page 327-390"},
		  	{dow:"T",date:"Dec 26th 2017", hw:"Questions 1-17"},
		  	{dow:"W",date:"Dec 27th 2017", hw:"HW Packet Page 1"},
		  	{dow:"T",date:"Dec 28th 2017", hw:"HW Packet Page 2"},
		  	{dow:"F",date:"Dec 29th 2017", hw:"HW Packet Page 3"},
		  	{dow:"S",date:"Dec 30th 2017", hw:"Sleep"},
		  	{dow:"S",date:"Dec 31st 2017", hw:"Sleep"},
		  	{dow:"M",date:"Jan 1st 2018", hw:"HW Packet Page 4"},
		  	{dow:"T",date:"Jan 2nd 2018", hw:"HW Packet Page 5-6"},
		  	{dow:"W",date:"Jan 3rd 2018", hw:"Pages 5-17"},
		  	{dow:"T",date:"Jan 4th 2018", hw:"Pages 18-27"},
		  	{dow:"F",date:"Jan 5th 2018", hw:"Pages 28-36"},
		  	{dow:"S",date:"Jan 6th 2018", hw:"Sleep"},
		  	{dow:"S",date:"Jan 7th 2018", hw:"Sleep"},
	*/
	Dow  string `json:"dow"`
	Date string `json:"date"`
	Hw   string `json:"hw"`
}

func calendarHandler(w http.ResponseWriter, r *http.Request) {
	data := []events{
		events{
			Dow:  "S",
			Date: "Dec 24th 2017",
			Hw:   "Sleep",
		},
		events{
			Dow:  "M",
			Date: "Dec 25th 2017",
			Hw:   "Chapter 8 page 327-390",
		},
		events{
			Dow:  "T",
			Date: "Dec 26th 2017",
			Hw:   "Questions 1-17",
		},
		events{
			Dow:  "W",
			Date: "Dec 27th 2017",
			Hw:   "HW Packet Page 1",
		},
		events{
			Dow:  "T",
			Date: "Dec 28th 2017",
			Hw:   "HW Packet Page 2",
		},
		events{
			Dow:  "F",
			Date: "Dec 29th 2017",
			Hw:   "HW Packet Page 3",
		},
		events{
			Dow:  "S",
			Date: "Dec 30th 2017",
			Hw:   "Sleep",
		},
		events{
			Dow:  "S",
			Date: "Dec 31st 2017",
			Hw:   "Sleep",
		},
		events{
			Dow:  "M",
			Date: "Jan 1st 2018",
			Hw:   "HW Packet Page 4",
		},
		events{
			Dow:  "T",
			Date: "Jan 2nd 2018",
			Hw:   "HW Packet Page 5-6",
		},
		events{
			Dow:  "W",
			Date: "Jan 3rd 2018",
			Hw:   "Pages 5-17",
		},
		events{
			Dow:  "T",
			Date: "Jan 4th 2018",
			Hw:   "Pages 18-27",
		},
		events{
			Dow:  "F",
			Date: "Jan 5th 2018",
			Hw:   "Pages 28-36",
		},
		events{
			Dow:  "S",
			Date: "Jan 6th 2018",
			Hw:   "Sleep",
		},
		events{
			Dow:  "S",
			Date: "Jan 7th 2018",
			Hw:   "Sleep",
		},
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		glog.Errorf("Failed to encode: %s", err)
	}
}

func main() {
	flag.Parse()

	// Resources are served directly.
	router := mux.NewRouter()
	router.PathPrefix("/res/").HandlerFunc(makeResourceHandler())

	// Add page handlers here.
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/classes", classesHandler)
	router.HandleFunc("/chat", chatHandler)
	router.HandleFunc("/calendar", calendarHandler)

	http.Handle("/", router)
	glog.Infof("Server is running at: http://localhost%s", *port)
	glog.Fatal(http.ListenAndServe(*port, nil))
}

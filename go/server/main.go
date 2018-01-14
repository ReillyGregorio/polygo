package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	firebase "firebase.google.com/go"
	"github.com/ReillyGregorio/polygo/go/ds"
	"github.com/gorilla/mux"
	"go.skia.org/infra/go/common"
	"go.skia.org/infra/go/sklog"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// flags
var (
	port         = flag.String("port", ":8000", "HTTP service address (e.g., ':8000')")
	resourcesDir = flag.String("resources_dir", "", "The directory to find templates, JS, and CSS files. If blank the current directory will be used.")
	local        = flag.Bool("local", true, "Running locally, as opposed to in production.")
)

var (
	templates   *template.Template
	firebaseApp *firebase.App
)

const (
	CLASSES  ds.Kind = "classes"
	CALENDAR ds.Kind = "calendar"
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
	sklog.Infof("index.html")
	if *local {
		loadTemplates()
	}
	data := IndexData{
		Name: "Skeleton App",
	}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		sklog.Errorf("Failed to expand template: %s", err)
	}
}

type period struct {
	/*
		{period:"1st",class:"Math",classroom:"Room 511"},
		{period:"2nd",class:"CSP",classroom:"Room 2703"},
		{period:"3rd",class:"English",classroom:"Room 2305"},
		{period:"4th",class:"Civics",classroom:"Room 1711"},
		{period:"5th",class:"Homeroom",classroom:"Room 503"}
	*/
	Period    string `json:"period"     datastore:"period"`
	Class     string `json:"class"      datastore:"class"`
	Classroom string `json:"classroom"  datastore:"classroom"`
	Semester  string `json:"semester"   datastore:"semester"`
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
		sklog.Errorf("Failed to encode: %s", err)
	}
}

type messages struct {
	/*
			{name:"Reilly:",msg:"What was the homework yesterday"},
		  	{name:"Isiah:",msg:"pages 5-17 I think"},
		  	{name:"Reilly:",msg:"oh right thx"},
		  	{name:"Generic Freshman:",msg:"I like cheese"},
	*/
	Name string    `json:"name"`
	Msg  string    `json:"msg"`
	TS   time.Time `json:"ts"`
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	data := []messages{
		messages{
			Name: "Reilly:",
			Msg:  "What was the homework yesterday",
			TS:   time.Now().Add(-4 * time.Hour),
		},
		messages{
			Name: "Isiah:",
			Msg:  "pages 5-17 I think",
			TS:   time.Now().Add(-3 * time.Hour),
		},
		messages{
			Name: "Reilly:",
			Msg:  "oh right thx",
			TS:   time.Now().Add(-2 * time.Hour),
		},
		messages{
			Name: "Generic Freshman:",
			Msg:  "I like cheese",
			TS:   time.Now().Add(-1 * time.Hour),
		},
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		sklog.Errorf("Failed to encode: %s", err)
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
	Date string `json:"date"`
	Hw   string `json:"hw"`
}

func calendarHandler(w http.ResponseWriter, r *http.Request) {
	/*data := []events{
		events{
			Date: "2017-12-24",
			Hw:   "Sleep",
		},
		events{
			Date: "2017-12-25",
			Hw:   "Chapter 8 page 327-390",
		},
		events{
			Date: "2017-12-26",
			Hw:   "Questions 1-17",
		},
		events{
			Date: "2017-12-27",
			Hw:   "HW Packet Page 1",
		},
		events{
			Date: "2017-12-28",
			Hw:   "HW Packet Page 2",
		},
		events{
			Date: "2017-12-29",
			Hw:   "HW Packet Page 3",
		},
		events{
			Date: "2017-12-30",
			Hw:   "Sleep",
		},
		events{
			Date: "2017-12-31",
			Hw:   "Sleep",
		},
		events{
			Date: "2018-01-01",
			Hw:   "HW Packet Page 4",
		},
		events{
			Date: "2018-01-02",
			Hw:   "HW Packet Page 5-6",
		},
		events{
			Date: "2018-01-03",
			Hw:   "Pages 5-17",
		},
		events{
			Date: "2018-01-04",
			Hw:   "Pages 18-27",
		},
		events{
			Date: "2018-01-05",
			Hw:   "Pages 28-36",
		},
		events{
			Date: "2018-01-06",
			Hw:   "Sleep",
		},
		events{
			Date: "2018-01-07",
			Hw:   "Sleep",
		},
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		sklog.Errorf("Failed to encode: %s", err)
	}*/

	lookup := map[string]events{}
	start := time.Now()
	start = start.Add(-time.Hour * 24 * 5)
	keys := []string{}
	for i := 0; i < 30; i++ {
		d := start.Format("2006-01-02")
		lookup[d] = events{Date: d}
		keys = append(keys, d)

		start = start.Add(time.Hour * 24)

	}
	query := ds.NewQuery(CALENDAR).Order("Date")
	data := []events{}
	it := ds.DS.Run(r.Context(), query)
	for {
		var e events
		_, err := it.Next(&e)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching next task: %v", err)
		}
		lookup[e.Date] = e
	}
	for _, key := range keys {
		data = append(data, lookup[key])
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		sklog.Errorf("Failed to encode: %s", err)
	}
}

type User struct {
	ID   string
	Name string
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sklog.Errorf("it errored %s", err)
		return
	}
	client, err := firebaseApp.Auth(context.Background())
	if err != nil {
		sklog.Errorf("error getting Auth client: %v\n", err)
		return
	}

	token, err := client.VerifyIDToken(string(b))
	if err != nil {
		sklog.Errorf("error verifying ID token: %v\n", err)
		return
	}

	sklog.Infof("Verified ID token: %v\n", token)
}

func calEditHandler(w http.ResponseWriter, r *http.Request) {
	e := events{}
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	sklog.Infof("%#v", e)
	key := ds.NewKey(CALENDAR)
	key.Name = e.Date
	_, err := ds.DS.Put(context.Background(), key, &e)
	if err != nil {
		sklog.Errorf("failed to write %s", err)
	}
}

func main() {
	common.Init()
	ds.Init("ultra-syntax-689", "production")
	opt := option.WithCredentialsFile("firebase.json")
	var err error
	firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	// Resources are served directly.
	router := mux.NewRouter()
	router.PathPrefix("/res/").HandlerFunc(makeResourceHandler())

	// Add page handlers here.
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/classes", classesHandler)
	router.HandleFunc("/chat", chatHandler)
	router.HandleFunc("/calendar", calendarHandler)
	router.HandleFunc("/verify", verifyHandler)
	router.HandleFunc("/calEdit", calEditHandler)

	http.Handle("/", router)
	sklog.Infof("Server is running at: http://localhost%s", *port)
	sklog.Fatal(http.ListenAndServe(*port, nil))
}

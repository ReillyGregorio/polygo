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
	"sort"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
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
	SCHEDULE ds.Kind = "schedule"
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
func classListHandler(w http.ResponseWriter, r *http.Request) {
	period := r.FormValue("period")
	semester := r.FormValue("semester")
	query := ds.NewQuery(CLASSES).Filter("period=", period).Filter("semester=", semester)
	data := []classes{}
	it := ds.DS.Run(r.Context(), query)
	for {
		var c classes
		_, err := it.Next(&c)
		if err == iterator.Done {
			break
		}
		if err != nil {
			sklog.Errorf("Error fetching next task: %v", err)
		}
		data = append(data, c)
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		sklog.Errorf("Failed to encode: %s", err)
	}
}

type classes struct {
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
type sliceOfClasses []*classes

func (a sliceOfClasses) Len() int           { return len(a) }
func (a sliceOfClasses) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sliceOfClasses) Less(i, j int) bool { return a[i].Period < a[j].Period }

type schedule struct {
	Classes []string `json:"classes"      datastore:"classes"`
}

func classesHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	semester := r.FormValue("semester")
	key := ds.NewKey(SCHEDULE)
	key.Name = uid + "-" + semester
	var s schedule
	if err := ds.DS.Get(r.Context(), key, &s); err != nil {
		sklog.Errorf("%v", err)
		http.Error(w, "not found 404", 404)
		return
	}
	dbkeys := []*datastore.Key{}
	for _, k := range s.Classes {
		key := ds.NewKey(CLASSES)
		key.Name = k
		dbkeys = append(dbkeys, key)
	}
	data := make([]*classes, len(dbkeys))
	if err := ds.DS.GetMulti(r.Context(), dbkeys, data); err != nil {
		sklog.Errorf("%v", err)
		http.Error(w, "not found", 404)
		return
	}
	allP := []string{"1st", "2nd", "3rd", "4th", "5th"}
	for _, p := range allP {
		found := false
		for _, c := range data {
			if c.Period == p {
				found = true
			}
		}
		if !found {
			data = append(data, &classes{
				Period:    p,
				Class:     "",
				Classroom: "",
				Semester:  semester,
			})
		}
	}
	sort.Sort(sliceOfClasses(data))
	/*data := []classes{
		{
			Period:    "1st",
			Class:     "Math",
			Classroom: "Room 511",
			Semester:  "2017-2",
		},
		{
			Period:    "2nd",
			Class:     "CSP",
			Classroom: "Room 2703",
			Semester:  "2017-2",
		},
		{
			Period:    "3rd",
			Class:     "English",
			Classroom: "Room 2305",
			Semester:  "2017-2",
		},
		{
			Period:    "4th",
			Class:     "Civics",
			Classroom: "Room 1711",
			Semester:  "2017-2",
		},
		{
			Period:    "5th",
			Class:     "Homeroom",
			Classroom: "Room 503",
			Semester:  "2017-2",
		},
	}*/
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
	Date     string `json:"date"`
	Hw       string `json:"hw"`
	Period   int    `json:"period"`
	Class    string `json:"class"`
	Semester string `json:"semester"`
}

func calendarHandler(w http.ResponseWriter, r *http.Request) {
	class := r.FormValue("class")
	semester := r.FormValue("semester")
	period, err := strconv.Atoi(r.FormValue("period"))

	if err != nil {
		sklog.Errorf("error in period conversion : %s", err)
		http.Error(w, "bad format", 400)
		return
	}
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
	query := ds.NewQuery(CALENDAR).Filter("Period=", period).Filter("Class=", class).Filter("Semester=", semester).Order("Date")
	data := []events{}
	it := ds.DS.Run(r.Context(), query)
	for {
		var e events
		_, err := it.Next(&e)
		if err == iterator.Done {
			break
		}
		if err != nil {
			sklog.Errorf("Error fetching next task: %v", err)
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

func isValidUser(r string) bool {
	client, err := firebaseApp.Auth(context.Background())
	if err != nil {
		sklog.Errorf("error getting Auth client: %v\n", err)
		return false
	}

	_, err = client.VerifyIDToken(r)
	if err != nil {
		sklog.Errorf("error verifying ID token: %v\n", err)
		return false
	}
	return true
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
	token := r.FormValue("token")
	if !isValidUser(token) {
		http.Error(w, "unauthorized", 402)
		return
	}
	e := events{}
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	sklog.Infof("%#v", e)
	key := ds.NewKey(CALENDAR)
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
	router.HandleFunc("/calendar", calendarHandler)
	router.HandleFunc("/verify", verifyHandler)
	router.HandleFunc("/calEdit", calEditHandler)
	router.HandleFunc("/classList", classListHandler)

	http.Handle("/", router)
	sklog.Infof("Server is running at: http://localhost%s", *port)
	sklog.Fatal(http.ListenAndServe(*port, nil))
}

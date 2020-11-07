package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type user struct {
	ID        int32  `json:"EMPLOYEE_INDEX"`
	Email     string `json:"EMAIL"`
	LastName  string `json:"LAST_NAME"`
	FirstName string `json:"FIRST_NAME"`
	JobTitle  string `json:"JOB_TITLE"`
	HireDate  string `json:"ADJ_HIRE_DATE"`
}

var (
	tpl                *template.Template
	db                 *sql.DB
	rwm                sync.RWMutex
	authenticatedUsers = make(map[string]ADUserInfo)
	dbUsers            = map[int32]user{} // user ID, user
)

func init() {

	tpl = template.Must(template.New("").Funcs(template.FuncMap{
		"DateFmt": func(t time.Time) string { return t.Format("2006-01-02") },
		"TimeFmt": func(t time.Time) string { return t.Format("15:04:05") },
	}).ParseGlob("templates/*"))

	// employee db (path to json file)
	employeeDb := os.Getenv("FORM503EMPLDB")
	if employeeDb == "" {
		employeeDb = "./data/employees.json"
	}
	loadDBFromJSON(employeeDb)
}

func main() {
	host := os.Getenv("FORM503HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("FORM503PORT")
	if port == "" {
		port = "8000"
	}

	dbfile := os.Getenv("FORM503DB")
	if dbfile == "" {
		dbfile = "./db.sqlite"
	}
	db = initDB(dbfile)

	log.Output(1, "starting server on "+host+":"+port)
	server := http.Server{Addr: host + ":" + port}

	http.HandleFunc("/", RequireAuth(index))
	http.HandleFunc("/confirm", confirm)
	http.HandleFunc("/data", data)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	if err := server.ListenAndServe(); err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
}

// RequireAuth - making sure user is authenticated
func RequireAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()

		fmt.Println("\tok:", ok)
		fmt.Println("\tu:", username)

		if !ok || !CheckUsernameAndPassword(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm=Login required`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised\n"))
			return
		}

		h.ServeHTTP(w, req)
		//h(w, req)
	}
}

func index(w http.ResponseWriter, req *http.Request) {

	var u user
	errstr := "Invalid request"

	username, _, ok := req.BasicAuth()
	if ok {
		uid := authenticatedUsers[username].ID
		iUID, _ := strconv.Atoi(uid)
		u = dbUsers[int32(iUID)]
		if u.ID > 0 {
			errstr = ""
		}
	}

	now := time.Now().Local()

	// process form submission
	if req.Method == http.MethodPost && errstr == "" {
		req.ParseForm()

		choice := req.PostForm.Get("choice")
		fmt.Println("got answer:", choice)

		if choice == "" {
			errstr = "Please choose an option."
		}

		if errstr == "" {
			answer := Answer{
				ID:       int(u.ID),
				FullName: u.FirstName + " " + u.LastName,
				Choice:   choice,
				DateTime: now, // Format("2006-01-02 15:04:05")
			}

			fmt.Println("about to store this data:", answer)
			//err := saveAnswer(answer, u)
			err := storeAnswer(db, answer)
			if err != nil {
				errstr = "storing answer: " + err.Error()
			}
		}

		if errstr == "" {
			w.Header().Add("Content-type", "text/plain")
			http.Redirect(w, req, "/confirm", 302)
			return
		}
	}

	todayStr := now.Format("1/2/2006")
	respData := struct {
		User   user
		ErrMsg string
		Today  string
	}{User: u, ErrMsg: errstr, Today: todayStr}

	err := tpl.ExecuteTemplate(w, "index.gohtml", respData)
	if err != nil {
		log.Println("Err: ", err)
	}
}

func confirm(w http.ResponseWriter, req *http.Request) {
	err := tpl.ExecuteTemplate(w, "confirm.gohtml", nil)
	if err != nil {
		log.Println("Err: ", err)
	}
}

func data(w http.ResponseWriter, req *http.Request) {

	query := req.URL.Query()
	date := query.Get("date")
	if date != "" {
		log.Printf("Got date: %+v", date)
	}

	type output struct{ Answers []Answer }
	var data output
	var err error
	answers, err := retrieveAnswers(db, date)
	if err != nil {
		fmt.Println(err)
	}
	data = output{Answers: answers}
	err = tpl.ExecuteTemplate(w, "data.gotxt", data)
	if err != nil {
		log.Println("Err: ", err)
	}
}

// loadDBFromJson - load json data in to dbUsers map
func loadDBFromJSON(dbPath string) {
	// read file
	data, err := ioutil.ReadFile(dbPath)
	if err != nil {
		fmt.Print(err)
		return
	}
	// json data
	var users []user
	// unmarshall it
	err = json.Unmarshal(data, &users)
	if err != nil {
		fmt.Println("error:", err)
	}

	for _, u := range users {
		// let's reformat the date
		hd, err := time.Parse("2006-01-02T00:00:00", u.HireDate)
		if err != nil {
			log.Println("+ Error parsing date: ", u.HireDate, err)
		} else {
			u.HireDate = hd.Format("1/2/2006")
		}
		dbUsers[u.ID] = u
	}

}

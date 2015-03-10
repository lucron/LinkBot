package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/thoas/stats"
)

func StartHttp2() {
	mux := http.NewServeMux()
	// routes
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/static/", staticHandler)

	http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", mux)
}

func StartHttp() {
	middleware := stats.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(middleware.Data())
		w.Write(b)
	})

	handler := middleware.Handler(mux)
	go http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 301)) // http -> https redirect
	http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", handler)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t := Template{}
	t.Load("index.html")

	var links string = ""

	db := Db{}
	db.Open()
	err := db.Query("select id, user, url, time from links order by id desc")
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	defer db.Close()

	for db.ResultRows.Next() {
		var id int64
		var user, url string
		var timestamp time.Time
		err = db.ResultRows.Scan(&id, &user, &url, &timestamp)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		links += fmt.Sprintf("<li class='lstItem'>%d. <div class='lstUrl'><a href='%s'>%s</a></div><div class='lstMeta'>von %s am %s</div></li>", id, url, url, user, timestamp.Format("02.01.2006 15:04"))
	}

	t.SetValue("{{lst_Links}}", links)
	io.WriteString(w, t.String())
}

// Handler for static css/js files
func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/"+r.URL.Path[1:])
}

func statsHandler(w http.ResponseWriter, r *http.Request) {

}

package main

import (
	"bufio"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	gocamole "github.com/jason0w0/gocamole/libs"
	template "github.com/jason0w0/gocamole/templates"
)

var db *gocamole.Config

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", templ.Handler(template.Home()).ServeHTTP)
	r.Post("/connection", ConnectionPage)
	r.Get("/ws/connect", WSConection)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("/app/static"))))

	http.ListenAndServe(":3000", r)
}

func ConnectionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}

	var ignoreCert string
	if r.FormValue("ignoreCertificate") == "on" {
		ignoreCert = "true"
	} else {
		ignoreCert = "false"
	}

	db = &gocamole.Config{
		Protocol:   r.FormValue("protocol"),
		Hostname:   r.FormValue("hostname"),
		Port:       r.FormValue("port"),
		Username:   r.FormValue("username"),
		Password:   r.FormValue("password"),
		Security:   r.FormValue("security"),
		IgnoreCert: ignoreCert,
	}

	template.Connection().Render(r.Context(), w)
}

func WSConection(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	upgrader.Subprotocols = append(upgrader.Subprotocols, "guacamole")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer ws.Close()

	conn, err := gocamole.InitializeGuacdConnection("guacd", 4822)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	config := db
	query := r.URL.Query()
	config.Screen = gocamole.Screen{
		Heigth: query.Get("height"),
		Width:  query.Get("width"),
		Dpi:    "96",
	}

	reader := bufio.NewReader(conn)
	if err := gocamole.StartHandshake(conn, ws, reader, config); err != nil {
		log.Println(err)
		return
	}

	go gocamole.WriteToGuacd(conn, ws)
	gocamole.ReadFromGuacd(conn, ws, reader)
}

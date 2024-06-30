package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/Naumovets/go-sfu/internal/sfu"
)

var (
	addr            = flag.String("addr", ":8080", "http service address")
	upgradeTemplate = &template.Template{}
)

func main() {
	// Parse the flags passed to program
	flag.Parse()

	// Init other state
	log.SetFlags(0)

	// Read upgrade.html from disk into memory, serve whenever anyone requests /
	upgradeHTML, err := os.ReadFile("upgrade.html")
	if err != nil {
		panic(err)
	}
	upgradeTemplate = template.Must(template.New("").Parse(string(upgradeHTML)))

	// websocket handler
	http.HandleFunc("/websocket", sfu.WebsocketHandler)

	// upgrade.html handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := upgradeTemplate.Execute(w, "ws://"+r.Host+"/websocket"); err != nil {
			log.Fatal(err)
		}
	})

	// request a keyframe every 3 seconds
	// go func() {
	// 	for range time.NewTicker(time.Second * 3).C {
	// 		sfu.DispatchKeyFrame("room")
	// 	}
	// }()

	// start HTTP server
	log.Fatal(http.ListenAndServe(*addr, nil)) // nolint:gosec
}

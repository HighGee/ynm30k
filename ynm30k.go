package main

import (
	"github.com/gorilla/mux"
)

func main() {
	s := new(Server)

	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
	r.HandleFunc("/trace/", s.TraceHandler)
	r.HandleFunc("/static/{file:.*}", s.FileHandler)
	r.HandleFunc("/dynamic/{file:.*}", s.DynamicHandler)
	r.HandleFunc("/code/{code:\\d{3}}", s.CodeHandler)
	r.HandleFunc("/size/{size:\\d+[km]{1}\\.(?:bin|zip)}", s.SizeHandler)
	r.HandleFunc("/headersize/{size:\\d+[km]{1}}", s.HeaderSizeHandler)
	r.HandleFunc("/slow/{range:\\d+-?(?:\\d+)?}", s.SlowHandler)
	r.HandleFunc("/redirect/{method:.*}", s.RedirectHandler)

	s.Init("0.0.0.0:80", r)

	s.Run()
}

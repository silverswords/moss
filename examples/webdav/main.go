package main

import (
	"log"
	"net/http"

	"golang.org/x/net/webdav"
)

func main() {
	srv := &webdav.Handler{
		FileSystem: webdav.Dir("."),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}

	http.Handle("/", srv)

	if err := http.ListenAndServe(":3100", nil); err != nil {
		log.Fatal(err)
	}
}

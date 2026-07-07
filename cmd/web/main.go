package main

import (
	"flag"
	"log"

	"github.com/maskedsyntax/jvman/internal/web"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	if err := web.New(*addr).ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

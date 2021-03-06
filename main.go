package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var (
	templatePath = flag.String("template", "./index.html", "path to template file")
	addr         = flag.String("addr", "127.0.0.1:5000", "listen address")
	factsPath    = flag.String("facts", "./facts.txt", "path to facts file")

	facts [][]byte
	tmpl  *template.Template
)

func readFacts() ([][]byte, error) {
	facts, err := ioutil.ReadFile(*factsPath)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(facts, []byte("\n"))
	for i, line := range lines {
		lines[i] = []byte(hex.Dump(line))
	}

	return lines, nil
}

func main() {
	flag.Parse()

	var err error
	facts, err = readFacts()
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err = tmpl.ParseFiles(*templatePath)
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().Unix())
	http.HandleFunc("/", routeRequest)
	log.Println("starting shark-hexdump")
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func routeRequest(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		serveFact(w, r)
	case "/upload":
		dumpBody(w, r)
	default:
		w.WriteHeader(404)
		io.WriteString(w, "not found")
	}
}

func serveFact(w http.ResponseWriter, r *http.Request) {
	fact := facts[rand.Intn(len(facts))]

	if strings.HasPrefix(r.UserAgent(), "curl") {
		w.Write(fact)
	} else {
		err := tmpl.Execute(w, string(fact))
		if err != nil {
			w.WriteHeader(500)
			io.WriteString(w, "internal server error")
			log.Println(err)
		}
	}
}

func dumpBody(w http.ResponseWriter, r *http.Request) {
	dumper := hex.Dumper(w)

	defer dumper.Close()
	_, err := io.Copy(dumper, r.Body)

	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "internal server error")
		log.Println(err)
	}
}

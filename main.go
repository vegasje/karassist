package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vegasje/karassist/ordered"
)

var songs = ordered.NewStringStringMap()
var queue []string
var mutex = &sync.Mutex{}

func main() {
	port := flag.String("p", "9080", "server port")
	flag.Parse()
	root := flag.Arg(0)
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".mp3") {
			sum := sha1.Sum([]byte(path))
			hash := base64.URLEncoding.EncodeToString(sum[:])
			songs.Set(hash, strings.TrimSuffix(f.Name(), ".mp3"))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Unable to walk directory: %s\n", err)
	}
	songs.SortByValue()
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/queue", queueHandler)
	mux.HandleFunc("/unqueue", unqueueHandler)
	log.Printf("Server started on port [%s]\n", *port)
	log.Fatalf("Unable to start server: %s\n", http.ListenAndServe(":"+*port, mux))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	search := r.URL.Query().Get("search")
	data := struct {
		Search string
		Queued string
		Songs  *ordered.StringStringMap
		Queue  []string
	}{
		Search: search,
		Queued: r.URL.Query().Get("queued"),
		Songs:  songs,
		Queue:  queue,
	}
	mutex.Unlock()
	if search != "" {
		data.Songs = songs.Filter(func(key, value string) bool {
			return CaseInsensitiveContains(value, search)
		})
	}
	t := template.Must(template.ParseFiles("views/index.tmpl"))
	if err := t.ExecuteTemplate(w, "main", data); err != nil {
		log.Printf("Error serving template: %s\n", err)
	}
}

func queueHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	mutex.Lock()
	queue = append(queue, id)
	mutex.Unlock()
	http.Redirect(w, r, "/?queued="+id, http.StatusFound)
}

func unqueueHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	for i := range queue {
		if queue[i] == id {
			mutex.Lock()
			queue = append(queue[:i], queue[i+1:]...)
			mutex.Unlock()
			break
		}
	}
	http.Redirect(w, r, "/?unqueued="+id, http.StatusFound)
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

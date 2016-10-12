package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vegasje/karassist/ordered"
)

var vlc string
var root string
var songs = ordered.NewStringStringMap()
var chQueue chan string
var chUnqueue chan string
var chGetQueue chan chan<- []string
var chPopQueue chan chan<- string

func main() {
	chQueue = make(chan string)
	chUnqueue = make(chan string)
	chGetQueue = make(chan chan<- []string)
	chPopQueue = make(chan chan<- string)

	port := flag.String("p", "9080", "server port")
	flag.StringVar(&vlc, "vlc", "/Applications/VLC.app/Contents/MacOS/VLC", "VLC location")
	flag.Parse()
	root = flag.Arg(0)

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
	go monitorQueue()
	go playQueue()
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/queue", queueHandler)
	mux.HandleFunc("/unqueue", unqueueHandler)
	log.Printf("Server started on port [%s]\n", *port)
	log.Fatalf("Unable to start server: %s\n", http.ListenAndServe(":"+*port, mux))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	ch := make(chan []string)
	chGetQueue <- ch
	data := struct {
		Search string
		Queued string
		Songs  *ordered.StringStringMap
		Queue  []string
	}{
		Search: search,
		Queued: r.URL.Query().Get("queued"),
		Songs:  songs,
		Queue:  <-ch,
	}
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
	chQueue <- id
	http.Redirect(w, r, "/?queued="+id, http.StatusFound)
}

func unqueueHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	chUnqueue <- id
	http.Redirect(w, r, "/?unqueued="+id, http.StatusFound)
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func monitorQueue() {
	var queue []string
	for {
		select {
		case id := <-chQueue:
			log.Printf("Queuing: %s\n", id)
			queue = append(queue, id)
		case id := <-chUnqueue:
			for i := range queue {
				if queue[i] == id {
					log.Printf("Unqueuing: %s\n", id)
					queue = append(queue[:i], queue[i+1:]...)
					break
				}
			}
		case ch := <-chGetQueue:
			log.Printf("Getting queue\n")
			ch <- queue
		case ch := <-chPopQueue:
			if len(queue) > 0 {
				log.Printf("Popping from queue: %s\n", queue[0])
				ch <- queue[0]
				queue = queue[1:]
			} else {
				ch <- ""
			}
		}
	}
}

func playQueue() {
	for {
		ch := make(chan string)
		chPopQueue <- ch
		filename := <-ch
		if filename != "" {
			log.Printf("Grabbed from queue: %s\n", filename)
			cmd := exec.Command(vlc, root+filename, "--video-on-top")
			cmd.Run()
		}
	}
}

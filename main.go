package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var songs sort.StringSlice

func main() {
	port := flag.String("p", "8080", "server port")
	flag.Parse()
	root := flag.Arg(0)
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".mp3") {
			songs = append(songs, strings.TrimSuffix(f.Name(), ".mp3"))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Unable to walk directory: %s\n", err)
	}
	songs.Sort()
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	log.Printf("Server started on port [%s]\n", *port)
	log.Fatalf("Unable to start server: %s\n", http.ListenAndServe(":"+*port, mux))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	data := struct {
		Search string
		Songs  []string
	}{
		Search: search,
		Songs:  songs,
	}
	if search != "" {
		data.Songs = Filter(songs, func(s string) bool {
			return CaseInsensitiveContains(s, search)
		})
	}
	t := template.Must(template.ParseFiles("views/index.tmpl"))
	if err := t.ExecuteTemplate(w, "main", data); err != nil {
		log.Printf("Error serving template: %s\n", err)
	}
}

func Filter(s []string, fn func(string) bool) []string {
	var p []string // == nil
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

// func view(resource string, w http.ResponseWriter, data interface{}) error {
// 	if _, ok := this.templates[resource]; !ok || !this.caching {
// 		this.templates[resource] = template.Must(template.ParseFiles("views/" + resource + ".tmpl"))
// 	}
// 	if err := this.templates[resource].ExecuteTemplate(w, "main", data); err != nil {
// 		return fmt.Errorf("Error serving template: %s", err)
// 	}
// 	return nil
// }

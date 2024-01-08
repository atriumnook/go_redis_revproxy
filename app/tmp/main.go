package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/redis/rueidis"
)

func main() {
	ctx := context.Background()
	redis, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{"redis:6379"}})
	if err != nil {
		log.Fatal("Redis Client: ", err)
	}
	defer redis.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path
		cmd := redis.B().Get().Key(key).Build()
		if val, err := redis.Do(ctx, cmd).ToString(); err != nil {
			log.Println("Redis Key: ", key, err)
			log.Println("Redis Do: ", err)
		} else {
			// w.Write([]byte(val))
			log.Println("Redis Key: ", key)
			if targetURL, err := url.Parse(val); err != nil {
				log.Println("URL Parse: ", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				log.Println("URL: ", targetURL)
				proxy := httputil.NewSingleHostReverseProxy(targetURL)
				originDirector := proxy.Director
				proxy.Director = func(r *http.Request) {
					originDirector(r)
					// r.URL.Scheme = targetURL.Scheme
					// r.URL.Host = targetURL.Host
					r.URL.Path = targetURL.Path
				}

				log.Println("Proxying To: ", targetURL)
				proxy.ServeHTTP(w, r)
			}
		}
	})

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/favicon.ico")
	})

	// テスト用
	http.HandleFunc("/image.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/image.png")
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

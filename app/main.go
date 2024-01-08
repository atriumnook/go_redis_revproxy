package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/rueidis"
)

var (
	redisClient rueidis.Client
	localCache  sync.Map
)

func initRedisClient() {
	var err error
	redisClient, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"redis:6379"},
		// コネクションプールの設定など
	})
	if err != nil {
		log.Fatal("Redis Client: ", err)
	}
}

func main() {
	initRedisClient()
	defer redisClient.Close()

	router := gin.Default()

	// ログ記録のためのミドルウェア
	router.Use(logMiddleware)

	// ルーティングの設定
	router.GET("/favicon.ico", serveFavicon)
	router.GET("/image.png", serveImage)
	// router.GET("/*path", handleRequest)
	router.NoRoute(handleRequest)

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Router Run: ", err)
	}
}

func logMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	duration := time.Since(start)
	log.Printf("[%s] %s %s %v", c.Request.Method, c.Request.URL.Path, c.ClientIP(), duration)
}

func handleRequest(c *gin.Context) {
	key := c.Request.URL.Path
	log.Println("Redis Key: ", key)
	cmd := redisClient.B().Get().Key(key).Build()
	val, err := redisClient.Do(context.Background(), cmd).ToString()
	log.Println("Redis Val: ", val)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		log.Println("Redis Key: ", key, err)
		return
	}

	serveProxy(c.Writer, c.Request, val)
}

// func serveProxy(w http.ResponseWriter, r *http.Request, target string) {
// 	targetURL, err := url.Parse(target)
// 	log.Println("URL: ", targetURL)
// 	if err != nil {
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		log.Println("URL Parse: ", err)
// 		return
// 	}
// 	newReverseProxy(targetURL)
// 	log.Println("newReverseProxyURL: ", targetURL)
// }

func serveProxy(w http.ResponseWriter, r *http.Request, target string) *httputil.ReverseProxy {
	parsedURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("URL Parse: ", err)
		return nil
	}
	log.Println("parsedURL: ", parsedURL)
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	// originalDirector := proxy.Director
	// proxy.Director = func(r *http.Request) {
	// 	originalDirector(r)
	// 	r.URL.Scheme = parsedURL.Scheme // スキーマの設定
	// 	r.URL.Host = parsedURL.Host     // ホストの設定
	// 	r.URL.Path = parsedURL.Path     // パスの設定
	// 	// 必要に応じて他のリクエストの属性も変更可能
	// }

	log.Println("proxy: ", proxy)
	return proxy
}

// func serveProxy(w http.ResponseWriter, r *http.Request, target string) {
// 	targetURL, err := url.Parse(target)
// 	log.Println("URL: ", targetURL)
// 	if err != nil {
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		log.Println("URL Parse: ", err)
// 		return
// 	}
// 	newReverseProxy(targetURL)
// 	log.Println("newReverseProxyURL: ", targetURL)
// }

// func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
// 	proxy := httputil.NewSingleHostReverseProxy(target)
// 	originalDirector := proxy.Director
// 	proxy.Director = func(req *http.Request) {
// 		originalDirector(req)
// 		req.URL.Scheme = target.Scheme // スキーマの設定
// 		req.URL.Host = target.Host     // ホストの設定
// 		req.URL.Path = target.Path     // パスの設定
// 		// 必要に応じて他のリクエストの属性も変更可能
// 	}
// 	return proxy
// }

func serveFavicon(c *gin.Context) {
	c.File("assets/favicon.ico")
}

func serveImage(c *gin.Context) {
	c.File("assets/image.png")
}

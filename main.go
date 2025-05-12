package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sync"

	"github.com/gocolly/colly"
)

type Listing struct {
	Image, Price, Name string
}

var listings []Listing
var mu sync.Mutex

func scrapeData() {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	c.OnHTML(".b-list-advert__gallery__item.js-advert-list-item", func(e *colly.HTMLElement) {
		listing := Listing{
			Image: e.ChildAttr("img", "src"),
			Price: e.ChildText(".qa-advert-price"),
			Name:  e.ChildText(".qa-advert-list-item-title .b-advert-title-inner"),
		}

		mu.Lock()
		listings = append(listings, listing)
		mu.Unlock()
	})

	c.Visit("https://jiji.co.ke/")
}

func main() {
	scrapeData()

	// Serve static files (e.g., CSS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, nil)
	})

	// Serve JSON API
	http.HandleFunc("/api/listings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		defer mu.Unlock()
		json.NewEncoder(w).Encode(listings)
	})

	http.ListenAndServe(":8081", nil)
}

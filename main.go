package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"golang.org/x/net/html"
)

var tpl = template.Must(template.ParseFiles("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

//Collect all links from response body and return it as an array of strings
func getLinks(body io.Reader) []string {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}

				}
			}

		}
	}
}

var test bool

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	fmt.Fprintf(w, "Information About Your URL \n")
	url := r.FormValue("url")
	fmt.Fprintf(w, "Url = %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	var metaDescription string
	var pageTitle string
	pageTitle = doc.Find("title").Contents().Text()
	// get title + description
	doc.Find("meta").Each(func(index int, item *goquery.Selection) {
		if item.AttrOr("name", "") == "description" {
			metaDescription = item.AttrOr("content", "")
		}
	})
	fmt.Fprintf(w, "Page Title: '%s'\n", pageTitle)
	fmt.Fprintf(w, "Meta Description: '%s'\n", metaDescription)

	// verify if there is a login form
	doc.Find("input[type='password']").Each(func(index int, element *goquery.Selection) {
		pss, exists := element.Attr("value")
		_ = pss
		if exists {
			test = true
		}

	})
	if test == true {
		fmt.Fprintf(w, "this page contains a login form \n")
	} else {
		fmt.Fprintf(w, "this page does not contains a login form \n")
	}
	//get Html version

	//test all links
	n := 0
	for _, v := range getLinks(resp.Body) {
		fmt.Println(v)
		resp, err := http.Get(v)
		_ = resp
		if err != nil {
			n++
			print(err.Error())
		}
	}
	fmt.Fprintf(w, "this page contains %d inaccessible links \n", n)
}

func main() {

	err := godotenv.Load()

	if err != nil {

		log.Println("Error loading .env file")

	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/informations", formHandler)
	http.ListenAndServe(":"+port, mux)
}

package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type informasiKuliah struct {
	kode  string
	nama  string
	kelas string
}

type ideData struct {
	np         string
	nama       string
	mataKuliah []string
}

func scrapIde(id string) ideData {
	c := colly.NewCollector(
		colly.AllowedDomains(
			"www.google.co.id",
			"www.google.com",
			"ide.unpar.ac.id",
		),
		colly.MaxDepth(1),
	)

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36"

	// logging
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println(err)
	})

	var result ideData

	// visit IDE link
	c.OnHTML("div.r", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a", "href")

		c.Visit(link)
	})

	c.OnHTML(".page-header-headings", func(e *colly.HTMLElement) {
		identifier := e.ChildText("h1")

		firstSpace := strings.Index(identifier, " ")
		result.np = identifier[0:firstSpace]
		result.nama = identifier[firstSpace+1:]
	})

	c.OnHTML(".node_category:nth-child(3) dd > ul > li", func(e *colly.HTMLElement) {
		text := e.ChildText("a")

		result.mataKuliah = append(result.mataKuliah, text)
	})

	url := fmt.Sprintf("https://www.google.co.id/search?ie=UTF-8&q=%s", id)

	c.Visit(url)

	return result
}

func main() {
	ideData := scrapIde("2017730017")

	fmt.Println(ideData.nama)
	fmt.Println(ideData.np)

	for _, text := range ideData.mataKuliah {
		fmt.Println(text)
	}
}

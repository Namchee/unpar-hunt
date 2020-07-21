package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type MataKuliah struct {
	Kode  string `json:"kode"`
	Nama  string `json:"nama"`
	Kelas string `json:"kelas"`
}

type RiwayatTOEFL struct {
	Listening int `json:"listening"`
	Structure int `json:"structure"`
	Reading   int `json:"reading"`
	Total     int `json:"total"`
}

type Mahasiswa struct {
	Np           string         `json:"np"`
	Foto         string         `json:"photo"`
	Nama         string         `json:"nama"`
	MataKuliah   []MataKuliah   `json:"mataKuliah"`
	RiwayatTOEFL []RiwayatTOEFL `json:"riwayatTOEFL"`
}

func scrapIde(baseCollector colly.Collector, id string) Mahasiswa {
	ideCollector := baseCollector.Clone()

	ideCollector.AllowedDomains = []string{
		"www.google.co.id",
		"www.google.com",
		"ide.unpar.ac.id",
	}

	var result Mahasiswa

	// visit IDE link from Google
	ideCollector.OnHTML("div.r", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a", "href")

		ideCollector.Visit(link)
	})

	// Get mahasiswa picture
	ideCollector.OnHTML(".page-header-image", func(e *colly.HTMLElement) {
		picURL := e.ChildAttrs("a > img", "src")

		result.Foto = picURL[0]
	})

	// Get mahasiswa np and name
	ideCollector.OnHTML(".page-header-headings", func(e *colly.HTMLElement) {
		identifier := e.ChildText("h1")

		firstSpace := strings.Index(identifier, " ")
		result.Np = identifier[0:firstSpace]
		result.Nama = identifier[firstSpace+1:]
	})

	// Get current MK
	ideCollector.OnHTML(".node_category:nth-child(3) dd > ul > li", func(e *colly.HTMLElement) {
		text := e.ChildText("a")
		textSplit := strings.Split(text, " ")

		mataKuliah := MataKuliah{
			textSplit[0:1][0],
			strings.Join(textSplit[1:len(textSplit)-1], " "),
			textSplit[len(textSplit)-1:][0],
		}

		result.MataKuliah = append(result.MataKuliah, mataKuliah)
	})

	url := fmt.Sprintf("https://www.google.co.id/search?ie=UTF-8&q=%s", id)

	ideCollector.Visit(url)

	return result
}

func main() {
	// initalize base collector
	baseCollector := colly.NewCollector(
		colly.MaxDepth(1),
	)

	// masquerade as a normal user
	baseCollector.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36"

	baseCollector.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting: ", req.URL)
	})

	baseCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println(err)
	})

	ideData := scrapIde(*baseCollector, "2017730054")

	jsonForm, err := json.Marshal(ideData)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(jsonForm))
}

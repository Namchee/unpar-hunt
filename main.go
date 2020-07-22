package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type MataKuliah struct {
	Kode  string `json:"kode"`
	Nama  string `json:"nama"`
	Kelas string `json:"kelas"`
}

type RiwayatTOEFL struct {
	Listening int       `json:"listening"`
	Structure int       `json:"structure"`
	Reading   int       `json:"reading"`
	Total     int       `json:"total"`
	Tanggal   time.Time `json:"tanggal"`
}

type Mahasiswa struct {
	Npm          string         `json:"npm"`
	Foto         string         `json:"photo"`
	Nama         string         `json:"nama"`
	MataKuliah   []MataKuliah   `json:"mataKuliah"`
	RiwayatTOEFL []RiwayatTOEFL `json:"riwayatTOEFL"`
}

func scrapIde(baseCollector colly.Collector, id string) (Mahasiswa, error) {
	ideCollector := baseCollector.Clone()

	ideCollector.AllowedDomains = []string{
		"www.google.co.id",
		"ide.unpar.ac.id",
	}

	var err error
	var result Mahasiswa

	ideCollector.OnError(func(_ *colly.Response, erro error) {
		err = erro
	})

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
		result.Npm = identifier[0:firstSpace]
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

	return result, err
}

func scrapTOEFL(baseCollector colly.Collector, id string) ([]RiwayatTOEFL, error) {

	toeflCollector := baseCollector.Clone()

	toeflCollector.AllowedDomains = []string{
		"cdc.unpar.ac.id",
	}

	var err error
	var result []RiwayatTOEFL

	toeflCollector.OnHTML("table tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(idx int, el *colly.HTMLElement) {
			if idx > 0 {
				listening, errL := strconv.Atoi(el.ChildText("td:nth-child(2)"))
				structure, _ := strconv.Atoi(el.ChildText("td:nth-child(3)"))
				reading, _ := strconv.Atoi(el.ChildText("td:nth-child(4)"))
				total, _ := strconv.Atoi(el.ChildText("td:nth-child(5)"))
				date, _ := time.Parse("2 January 2006", el.ChildText("td:nth-child(6)"))

				if errL == nil {
					result = append(
						result,
						RiwayatTOEFL{
							listening,
							structure,
							reading,
							total,
							date,
						},
					)
				}
			}
		})
	})

	err = toeflCollector.Post(
		"http://cdc.unpar.ac.id/old/cek-nilai-toefl/",
		map[string]string{
			"npm": id,
			"act": "Cek Nilai",
		},
	)

	return result, err
}

func main() {
	// initalize base collector
	baseCollector := colly.NewCollector(
		colly.MaxDepth(1),
	)

	// masquerade as a normal user
	baseCollector.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36"

	ideData, err := scrapIde(*baseCollector, "2016730011")

	if err == nil && len(ideData.Npm) > 0 {
		toeflData, erro := scrapTOEFL(*baseCollector, "2016730011")

		if erro == nil {
			ideData.RiwayatTOEFL = toeflData
		}
	} else {
		fmt.Println(err)
		os.Exit(1)
	}

	jsonForm, err := json.Marshal(ideData)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(jsonForm))
}

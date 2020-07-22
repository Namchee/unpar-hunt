package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type mataKuliah struct {
	Kode  string `json:"kode"`
	Nama  string `json:"nama"`
	Kelas string `json:"kelas"`
}

type riwayatTOEFL struct {
	Listening int       `json:"listening"`
	Structure int       `json:"structure"`
	Reading   int       `json:"reading"`
	Total     int       `json:"total"`
	Tanggal   time.Time `json:"tanggal"`
}

type informasiIde struct {
	Npm        string       `json:"npm"`
	Foto       string       `json:"photo"`
	Nama       string       `json:"nama"`
	MataKuliah []mataKuliah `json:"mataKuliah"`
}

type Mahasiswa struct {
	informasiIde
	RiwayatTOEFL []riwayatTOEFL `json:"riwayatTOEFL"`
}

// Scrap basic student data from IDE
func scrapIde(baseCollector colly.Collector, id string) (informasiIde, error) {
	ideCollector := baseCollector.Clone()

	ideCollector.AllowedDomains = []string{
		"www.google.co.id",
		"ide.unpar.ac.id",
	}

	var err error
	var result informasiIde

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

		mataKuliah := mataKuliah{
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

// Scrap TOEFL history from UNPAR's CDC website
func scrapTOEFL(baseCollector colly.Collector, id string) ([]riwayatTOEFL, error) {
	toeflCollector := baseCollector.Clone()

	toeflCollector.AllowedDomains = []string{
		"cdc.unpar.ac.id",
	}

	var err error
	var result []riwayatTOEFL

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
						riwayatTOEFL{
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

func ScrapData(id string) (Mahasiswa, error) {
	// initalize base collector
	baseCollector := colly.NewCollector(
		colly.MaxDepth(1),
	)

	// masquerade as a normal user
	baseCollector.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36"

	ideData, errIde := scrapIde(*baseCollector, id)
	toeflData, errToefl := scrapTOEFL(*baseCollector, id)

	if errIde != nil {
		return Mahasiswa{}, errIde
	}

	if errToefl != nil {
		return Mahasiswa{}, errToefl
	}

	return Mahasiswa{ideData, toeflData}, nil
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatal("IDK where the hell should I scrap the data from...")
	}

	data, errData := ScrapData(args[0])

	if errData != nil {
		log.Fatalln(errData)
	}

	jsonForm, err := json.MarshalIndent(data, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonForm))
}

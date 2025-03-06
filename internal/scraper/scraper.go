package scraper

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

func writeNews(writer *csv.Writer, newsList []*News) {
	for _, news := range newsList {
		err := writer.Write([]string{
			news.Title,
			news.Tag,
			news.Author,
			news.Date,
			news.Url,
			news.Image,
			news.Content,
		})
		if err != nil {
			fmt.Println("Error writing data to CSV")
		}
	}
}

func getTotalPages(url string) int {
	var lastPage string

	pageCollector := colly.NewCollector()

	pageCollector.OnHTML("nav.pagination div.nav-links", func(e *colly.HTMLElement) {
		lastPage = e.ChildText("a.page-numbers:nth-last-child(2)")
	})

	pageCollector.Visit(url)

	if lastPage == "" {
		return 1
	}

	// Last page can be in "2,345" format, including a ","
	lastPageParsed := strings.ReplaceAll(lastPage, ",", "")
	num, err := strconv.Atoi(lastPageParsed)
	if err != nil {
		return 1
	}

	return num
}

func scrapeNews(wg *sync.WaitGroup, url string, category string, ch chan<- []*News) {
	defer wg.Done()

	var newsList []*News

	c := colly.NewCollector(
		colly.AllowedDomains("guerrero.quadratin.com.mx"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*.quadratin.com.mx",
		Parallelism: 2,
		Delay:       5 * time.Second,
	})

	c.SetRequestTimeout(60 * time.Second)

	detailCollector := c.Clone()

	c.OnHTML("section.q-main-component", func(e *colly.HTMLElement) {
		e.ForEach("article.q-notice", func(_ int, news *colly.HTMLElement) {
			link := news.ChildAttr("a:nth-child(2)", "href")
			tag := news.ChildText("div.tag")
			parsedTag := parseSpanishWord(tag)
			if parsedTag == category {
				detailCollector.Visit(link)
			}
		})
	})

	detailCollector.OnHTML("div.q-content", func(e *colly.HTMLElement) {
		author := "-"
		redaction := e.ChildText("div.q-content__redacted")
		parts := strings.Split(redaction, "/")
		if len(parts) > 0 {
			author = parts[0]
		}

		date := e.ChildText("div.date") + " " + e.ChildText("div.hour")

		content := ""
		e.ForEach("div p", func(_ int, p *colly.HTMLElement) {
			content += p.Text + "\n"
		})

		news := &News{
			Title:   e.ChildText("h1"),
			Tag:     category,
			Author:  author,
			Date:    date,
			Url:     e.Request.URL.String(),
			Image:   e.ChildAttr("img", "src"),
			Content: content,
		}
		newsList = append(newsList, news)
		news.PrintNews()
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting Page: ", r.URL)
	})

	detailCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting News: ", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.Visit(url)
	c.Wait()
	detailCollector.Wait()

	ch <- newsList
}

func GetNews() {
	csvFile, err := os.Create("data/news.csv")
	if err != nil {
		fmt.Println("Error while creating file:", err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	csvWriter.Write([]string{
		"Title",
		"Tag",
		"Author",
		"Date",
		"Url",
		"Image",
		"Content",
	})

	var wg sync.WaitGroup
	ch := make(chan []*News, len(newsCategories))

	for _, category := range newsCategories {
		url := newsUrl + category + "/"
		totalPages := getTotalPages(url)
		for p := 1; p <= totalPages; p++ {
			wg.Add(1)
			pageUrl := url + "page/" + strconv.Itoa(p) + "/"
			go scrapeNews(&wg, pageUrl, category, ch)
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for newsList := range ch {
		writeNews(csvWriter, newsList)
	}
}

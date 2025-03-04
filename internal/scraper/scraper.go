package scrapper

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"

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

func scrapeNews(wg *sync.WaitGroup, category string, ch chan<- []*News) {
	defer wg.Done()

	var newsList []*News

	c := colly.NewCollector(
		colly.CacheDir("./quadrantin_cache"),
		colly.AllowedDomains("guerrero.quadratin.com.mx"),
	)

	c.OnHTML("article.q-notice a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})

	// c.OnHTML("a.next.page-numbers[href]", func(e *colly.HTMLElement) {
	// 	nextLink := e.Attr("href")
	// 	if nextLink != "" {
	// 		e.Request.Visit(nextLink)
	// 	}
	// })

	c.OnHTML("div.q-content", func(e *colly.HTMLElement) {
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
		// news.PrintNews()
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	url := newsUrl + category
	c.Visit(url)

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
		"Hour",
		"Image",
		"Content",
	})

	var wg sync.WaitGroup
	ch := make(chan []*News, len(newsCategories))

	for _, category := range newsCategories {
		wg.Add(1)
		go scrapeNews(&wg, category, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for newsList := range ch {
		writeNews(csvWriter, newsList)
	}
}

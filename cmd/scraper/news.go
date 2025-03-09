package main

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

	"github.com/victormhp/qudratin-gro/internal/models"
)

func scrapCategory(
	wg *sync.WaitGroup,
	category *models.Category,
	ch chan<- []*models.News,
	pages ...int,
) {
	defer wg.Done()

	var totalPages int
	categoryUrl := newsUrl + "category/" + category.Name + "/"

	c := colly.NewCollector(
		colly.AllowedDomains("guerrero.quadratin.com.mx"),
	)

	c.OnHTML("nav.pagination div.nav-links", func(e *colly.HTMLElement) {
		lastPage := e.ChildText("a.page-numbers:nth-last-child(2)")
		parsedLastPage, err := strconv.Atoi(strings.ReplaceAll(lastPage, ",", ""))
		if err != nil {
			fmt.Println("Failed to parse total pages")
		}
		totalPages = parsedLastPage
	})

	c.Visit(categoryUrl)

	var pageWg sync.WaitGroup
	limiter := make(chan struct{}, 2)

	if pages[0] != 0 {
		totalPages = pages[0]
	}
	for p := 1; p <= totalPages; p++ {
		pageWg.Add(1)
		limiter <- struct{}{}

		go func(pageNum int) {
			defer pageWg.Done()
			defer func() { <-limiter }()

			pageUrl := categoryUrl + "page/" + strconv.Itoa(pageNum) + "/"
			newsList := scrapeNews(pageUrl, category)

			if len(newsList) > 0 {
				ch <- newsList
			}

			time.Sleep(5 * time.Second)
		}(p)
	}

	pageWg.Wait()
	fmt.Println("Scraped all news from category:", category.Name)
}

func scrapeNews(url string, category *models.Category) []*models.News {
	var newsList []*models.News

	pageCollector := colly.NewCollector(
		colly.AllowedDomains("guerrero.quadratin.com.mx"),
		colly.Async(true),
	)

	pageCollector.Limit(&colly.LimitRule{
		DomainGlob:  "guerrero.quadratin.com.mx",
		Parallelism: 2,
		Delay:       5 * time.Second,
	})

	pageCollector.SetRequestTimeout(60 * time.Second)

	newsCollector := pageCollector.Clone()

	pageCollector.OnHTML("section.q-main-component", func(e *colly.HTMLElement) {
		e.ForEach("article.q-notice", func(_ int, news *colly.HTMLElement) {
			link := news.ChildAttr("a:nth-child(2)", "href")
			tag := news.ChildText("div.tag")
			parsedTag := parseSpanishWord(tag)
			if parsedTag == category.Name {
				newsCollector.Visit(link)
			}
		})
	})

	newsCollector.OnHTML("div.q-content", func(e *colly.HTMLElement) {
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
			content = strings.ReplaceAll(strings.TrimSpace(content), "\"", "'")
		})

		news := &models.News{
			CategoryId: category.Id,
			Title:      e.ChildText("h1"),
			Author:     author,
			Date:       date,
			Url:        e.Request.URL.String(),
			Image:      e.ChildAttr("img", "src"),
			Content:    content,
		}
		newsList = append(newsList, news)
	})

	pageCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting Page: ", r.URL)
	})

	newsCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting News: ", r.URL)
	})

	pageCollector.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	pageCollector.Visit(url)
	pageCollector.Wait()
	newsCollector.Wait()

	return newsList
}

func getNews(categories []*models.Category, pages ...int) []*models.News {
	var wg sync.WaitGroup
	ch := make(chan []*models.News, len(categories))

	page := 0
	if len(pages) > 0 {
		page = pages[0]
	}

	for _, category := range categories {
		wg.Add(1)
		go scrapCategory(&wg, category, ch, page)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var news []*models.News
	for newsList := range ch {
		news = append(news, newsList...)
	}

	return news
}

func writeNewsToCsv(news []*models.News) error {
	csvFile, err := os.Create("data/news.csv")
	if err != nil {
		return fmt.Errorf("Error while creating file: %w", err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{
		"id",
		"category_id",
		"title",
		"author",
		"date",
		"url",
		"image",
		"content",
	}); err != nil {
		return fmt.Errorf("Error writing header: %w", err)
	}

	for i, n := range news {
		if err := csvWriter.Write([]string{
			strconv.Itoa(i + 1),
			strconv.Itoa(n.CategoryId),
			n.Title,
			n.Author,
			n.Date,
			n.Url,
			n.Image,
			n.Content,
		}); err != nil {
			return fmt.Errorf("Error writing data to CSV %w", err)
		}
	}

	return nil
}

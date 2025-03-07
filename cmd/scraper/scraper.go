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

const newsUrl = "https://guerrero.quadratin.com.mx/"

func getCategories() []*models.Category {
	var categories []*models.Category

	c := colly.NewCollector(
		colly.AllowedDomains("guerrero.quadratin.com.mx"),
	)

	c.OnHTML("nav.q-menu ul", func(e *colly.HTMLElement) {
		e.ForEach("li:nth-child(n+3)", func(i int, a *colly.HTMLElement) {
			category := a.ChildText("a")
			parsedCategory := parseSpanishWord(category)
			newCategory := &models.Category{Id: i + 1, Name: parsedCategory}
			categories = append(categories, newCategory)
		})
	})

	c.Visit(newsUrl)

	return categories
}

func writeCategoriesToCsv(categories []*models.Category) error {
	csvFile, err := os.Create("data/categories.csv")
	if err != nil {
		return fmt.Errorf("Error while creating file: %w", err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{
		"id",
		"name",
	}); err != nil {
		return fmt.Errorf("Error writing header: %w", err)
	}

	for _, c := range categories {
		err := csvWriter.Write([]string{
			strconv.Itoa(c.Id),
			c.Name,
		})
		if err != nil {
			return fmt.Errorf("Error writing data to CSV: %w", err)
		}
	}

	return nil
}

func scrapCategory(
	wg *sync.WaitGroup,
	category *models.Category,
	ch chan<- []*models.News,
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
		})

		news := &models.News{
			Title:      e.ChildText("h1"),
			CategoryId: category.Id,
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

func getNews(categories []*models.Category) []*models.News {
	var wg sync.WaitGroup
	ch := make(chan []*models.News, len(categories))

	for _, category := range categories {
		wg.Add(1)
		go scrapCategory(&wg, category, ch)
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
		"CategoryId",
		"Title",
		"Author",
		"Date",
		"Url",
		"Image",
		"Content",
	}); err != nil {
		return fmt.Errorf("Error writing header: %w", err)
	}

	for _, n := range news {
		if err := csvWriter.Write([]string{
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

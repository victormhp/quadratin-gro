package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/gocolly/colly"

	"github.com/victormhp/qudratin-gro/internal/models"
)

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

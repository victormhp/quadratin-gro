package main

import (
	"fmt"

	"github.com/victormhp/qudratin-gro/internal/models"
)

func main() {
	var cs []*models.Category
	var c models.Category
	c.Id = 1
	c.Name = "politica"
	cs = append(cs, &c)

	news := getNews(cs)
	if err := writeNewsToCsv(news); err != nil {
		fmt.Println(err)
	}
}

package scrapper

import "fmt"

type News struct {
	Title   string
	Tag     string
	Author  string
	Date    string
	Url     string
	Image   string
	Content string
}

func (n *News) PrintNews() {
	fmt.Printf("Title: %s\n", n.Title)
	fmt.Printf("Tag: %s\n", n.Title)
	fmt.Printf("Url: %s\n", n.Url)
	fmt.Printf("Date: %s\n", n.Date)
	fmt.Printf("Image: %s\n", n.Image)
	fmt.Printf("Author: %s\n", n.Author)
	fmt.Printf("Content: %s\n", n.Content)
	fmt.Println()
}

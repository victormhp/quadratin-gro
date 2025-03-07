package models

import "fmt"

type News struct {
	Id         int
	CategoryId int
	Title      string
	Author     string
	Date       string
	Url        string
	Image      string
	Content    string
}

func (n *News) PrintNews() {
	fmt.Printf("Id: %d\n", n.Id)
	fmt.Printf("CategoryId: %d\n", n.Id)
	fmt.Printf("Title: %s\n", n.Title)
	fmt.Printf("Author: %s\n", n.Author)
	fmt.Printf("Date: %s\n", n.Date)
	fmt.Printf("Url: %s\n", n.Url)
	fmt.Printf("Image: %s\n", n.Image)
	fmt.Printf("Content: %s\n", n.Content)
	fmt.Println()
}

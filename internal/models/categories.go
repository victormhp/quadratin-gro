package models

import "fmt"

type Category struct {
	Id   int64    `json:"id"`
	Name string `json:"name"`
}

func (c *Category) PrintCategory() {
	fmt.Printf("Id: %d\n", c.Id)
	fmt.Printf("Tag: %s\n", c.Name)
	fmt.Println()
}

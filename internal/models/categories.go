package models

import "fmt"

type Category struct {
	Id   int
	Name string
}

func (c *Category) PrintCategory() {
	fmt.Printf("Id: %d\n", c.Id)
	fmt.Printf("Tag: %s\n", c.Name)
	fmt.Println()
}

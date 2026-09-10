package main

import (
	"fmt"

	ticketstore "tiyu/storage/pebble"
)

func main() {
	if err := ticketstore.ClearAll(); err != nil {
		panic(err)
	}
	fmt.Println("pbdata cleared")
}

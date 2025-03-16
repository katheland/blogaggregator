package main

import (
	"fmt"
	"internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	cfg.SetUser("kendrel")

	cfg2, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cfg2)
}
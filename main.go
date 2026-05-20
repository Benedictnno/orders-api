package main

import (
	"fmt"
	"context"

	"github.com/Benedictnno/orders-api/application"
)

func main() {
	app := application.New()
	err := app.Start(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

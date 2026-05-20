package handler

import (
	"net/http"
	"fmt"
)

type Order struct {}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Create order")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "List order")
}

func (o *Order) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Get order")
}
func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Get order by ID")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Update order by ID")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Delete order by ID")
}
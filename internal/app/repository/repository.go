package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Order struct {
	ID    int
	Title string
}

type Fuel struct {
	ID        int
	Title     string
	Heat      float64
	CardImage string
	ShortDesc string
	FullDesc  string
}

func (r *Repository) GetReqFuels() ([]Fuel, error) {
	// имитация получения списка id топлива в заявке
	reqs := []int{2, 4}
	var reqFuels []Fuel
	fuels, err := r.GetFuels()
	if err != nil {
		return nil, err
	}
	for _, id := range reqs {
		for _, fuel := range fuels {
			if fuel.ID == id {
				reqFuels = append(reqFuels, fuel)
			}
		}
	}
	return reqFuels, nil

}

func (r *Repository) GetFuels() ([]Fuel, error) {
	fuels := []Fuel{
		{
			ID:        1,
			Title:     "Метан",
			Heat:      50.1,
			CardImage: "http://127.0.0.1:9000/ripimages/metan.png",
			ShortDesc: "Безопасное, экологичное и экономичное моторное топливо," +
				" получаемое из природного газа",
			FullDesc: "Метан является экологически чистым, экономически выгодным" +
				" и безопасным топливом, получаемым из природного газа. Он" +
				" используется в качестве моторного топлива для транспорта " +
				"(компримированный природный газ - КПГ или CNG), а также для" +
				" отопления и приготовления пищи. Его преимущества включают низкие" +
				" выбросы вредных веществ, долговечность двигателя и низкую стоимость",
		},
		{
			ID:        2,
			Title:     "Пропан-бутан",
			Heat:      43.8,
			CardImage: "http://127.0.0.1:9000/ripimages/propanbutan.jpg",
			ShortDesc: "text",
			FullDesc:  "text",
		},
		{
			ID:        3,
			Title:     "Ацетилен",
			Heat:      50.4,
			CardImage: "http://127.0.0.1:9000/ripimages/ballongaz-acetilen1.png",
			ShortDesc: "text",
			FullDesc:  "text",
		},
		{
			ID:        4,
			Title:     "Водород",
			Heat:      141,
			CardImage: "http://127.0.0.1:9000/ripimages/hydrogen.jpg",
			ShortDesc: "text",
			FullDesc:  "text",
		},
		{
			ID:        5,
			Title:     "Дизельное топливо",
			Heat:      42.7,
			CardImage: "http://127.0.0.1:9000/ripimages/diesel.jpeg",
			ShortDesc: "text",
			FullDesc:  "text",
		},
	}
	if len(fuels) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return fuels, nil

}

func (r *Repository) GetOrders() ([]Order, error) {
	orders := []Order{
		{
			ID:    1,
			Title: "first order",
		},
		{
			ID:    2,
			Title: "second order",
		},
		{
			ID:    3,
			Title: "third order",
		},
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return orders, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil
		}
	}

	return Order{}, fmt.Errorf("заказ не найден")

}

func (r *Repository) GetFuel(id int) (Fuel, error) {
	fuels, err := r.GetFuels()
	if err != nil {
		return Fuel{}, err
	}

	for _, fuel := range fuels {
		if fuel.ID == id {
			return fuel, nil
		}
	}
	return Fuel{}, fmt.Errorf("отсутсвует топливо")
}

func (r *Repository) GetOrderByTitle(title string) ([]Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return []Order{}, err
	}

	var result []Order
	for _, order := range orders {
		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
			result = append(result, order)
		}
	}

	return result, nil
}

package main

import (
	"github.com/pariz/gountries"
)

func getStates() ListItems {
	var ll ListItems
	query := gountries.New()

	India, _ := query.FindCountryByName("india")
	subdivisions := India.SubDivisions()

	for _, subdivision := range subdivisions {
		ll = append(ll, ListItem{
			Label: subdivision.Name,
			Value: subdivision.Code,
		})
	}
	return ll
}

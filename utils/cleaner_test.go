package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeCounterparty(t *testing.T) {
	tests := []struct {
		rawName  string
		expected string
	}{
		{"ATB", "АТБ"},
		{"атб", "АТБ"},
		{"SILPO", "Сільпо"},
		{"CILPO", "Сільпо"},
		{"FORA", "Фора"},
		{"NOVUS", "Novus"},
		{"НОВУС", "Novus"},
		{"MCDONALDS", "McDonald's"},
		{"макдональдз", "McDonald's"},
		{"UKLON", "Uklon"},
		{"KLO", "KLO"},
		{"UBER", "Uber"},
		{"BOLT", "Bolt"},
		{"NOVAPAY", "NovaPay"},
		{"NOVA POSHTA", "Нова Пошта"},
		{"УКРПОШТА", "Укрпошта"},
		{"ROZETKA", "Rozetka"},
		{"Prom.ua", "Prom.ua"},
		{"ALIEXPRESS", "AliExpress"},
		{"WOG", "WOG"},
		{"OKKO", "OKKO"},
		{"EPITSENTR", "Епіцентр"},
		{"JYSK", "Jysk"},
		{"EVA", "Eva"},
		{"ANC", "Аптека АНЦ"},
		{"9-1-1", "Аптека 9-1-1"},
		{"Megogo", "Megogo"},
		{"Spotify", "Spotify"},
		{"NETFLIX", "Netflix"},
		{"YouTube Premium", "Youtube"},
		{"Readeat", "Readeat"},
		{"Книгарня Readeat", "Readeat"},
		// Trash cleaning tests
		{"FOP IVANOV", "IVANOV"},
		{"MAGAZYN PRODUKTY", "PRODUKTY"},
		{"ID платежу: 12345", ""},
		{"ABC* TRANSACTION", "TRANSACTION"},
		{"NAME, CITY", "NAME"},
		{"NAME 12345.67", "NAME"},
		{"  Spaces  ", "Spaces"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.rawName, func(t *testing.T) {
			result := NormalizeCounterparty(tt.rawName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

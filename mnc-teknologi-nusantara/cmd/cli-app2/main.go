package main

import (
	"fmt"
	"log"
	"math"
)

func main() {
	var totalBelanja, pembayaran float64

	// Prompt the user for the total shopping amount
	fmt.Print("Total belanja seorang customer: Rp ")
	_, err := fmt.Scan(&totalBelanja)
	if err != nil {
		log.Fatalf("Input error: %v", err)
	}

	// Prompt the user for the payment amount
	fmt.Print("Pembeli membayar: Rp ")
	_, err = fmt.Scan(&pembayaran)
	if err != nil {
		log.Fatalf("Input error: %v", err)
	}

	// Calculate the change
	change := pembayaran - totalBelanja

	// Check if the payment is sufficient
	if change < 0 {
		fmt.Println("False, Pembayaran kurang")
	} else {
		// Round the change to the nearest hundred
		roundedChange := math.Floor(change/100) * 100

		// Output the change
		fmt.Printf("Kembalian yang harus diberikan kasir: %.0f,\ndibulatkan menjadi %.0f\n", change, roundedChange)

		// Break down the change into denominations
		breakdownDenominations(roundedChange)
	}
}

func breakdownDenominations(amount float64) {
	denominations := []struct {
		value float64
		name  string
	}{
		{50000, "lembar 50.000"},
		{20000, "lembar 20.000"},
		{10000, "lembar 10.000"},
		{5000, "lembar 5.000"},
		{2000, "lembar 2.000"},
		{1000, "lembar 1.000"},
		{500, "koin 500"},
		{200, "koin 200"},
		{100, "koin 100"},
	}

	fmt.Println("Pecahan uang:")
	for _, denom := range denominations {
		count := int(amount / denom.value)
		if count > 0 {
			fmt.Printf("%d %s\n", count, denom.name)
			amount -= float64(count) * denom.value
		}
	}
}

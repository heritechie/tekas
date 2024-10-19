package main

import (
	"fmt"
	"log"
	"strings"
)

func main() {
	var n int

	// Read the number of input strings
	_, err := fmt.Scan(&n)
	if err != nil {
		log.Fatalf("Failed to read number of strings: %v", err)
	}

	// Read the input strings
	inputs := make([]string, n)
	for i := 0; i < n; i++ {
		_, err := fmt.Scan(&inputs[i])
		if err != nil {
			log.Fatalf("Failed to read input string %d: %v", i, err)
		}
	}

	mapMatches := make(map[string][]int)

	for i, input := range inputs {
		mapMatches[input] = append(mapMatches[input], i+1)
	}

	var output []string
	for _, mapInput := range mapMatches {
		if len(mapInput) > 1 {
			for _, index := range mapInput {
				output = append(output, fmt.Sprintf("%d", index))
			}
		}
	}

	// output
	fmt.Println(strings.Join(output, " "))
}

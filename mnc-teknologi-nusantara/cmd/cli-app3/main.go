package main

import (
	"fmt"
)

func isValid(input string) bool {
	// Validasi panjang string
	if len(input) < 1 || len(input) > 4096 {
		return false
	}

	// Stack untuk menyimpan karakter pembuka
	stack := []rune{}

	// Peta untuk mencocokkan pembuka dan penutup
	pairs := map[rune]rune{
		'{': '}',
		'[': ']',
		'<': '>',
		'(': ')',
	}

	// Memeriksa setiap karakter dalam input
	for _, char := range input {
		if _, ok := pairs[char]; ok { // Jika karakter adalah pembuka
			stack = append(stack, char)
		} else {
			// Jika karakter adalah penutup
			if len(stack) == 0 {
				return false // Penutup tanpa pembuka
			}
			// Mengambil karakter terakhir dari stack
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1] // Menghapus karakter terakhir dari stack
			if pairs[top] != char { // Memeriksa kecocokan
				return false
			}
		}

		// Memeriksa pengurungan yang salah
		if len(stack) > 1 {
			// Memeriksa apakah elemen kedua di dalam stack adalah penutup dari elemen pertama
			if pairs[stack[len(stack)-2]] == stack[len(stack)-1] {
				return false
			}
		}
	}

	// Jika stack kosong, berarti semua pembuka sudah ditutup
	return len(stack) == 0
}

func main() {
	var input string

	// Menerima input dari pengguna
	fmt.Print("Input: ")
	fmt.Scan(&input)

	// Memeriksa validitas input dan menampilkan hasil
	result := isValid(input)
	fmt.Printf("Output: %t\n", result)
}

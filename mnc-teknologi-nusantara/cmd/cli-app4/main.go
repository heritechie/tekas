package main

import (
	"fmt"
	"time"
    "strconv"
)

const (
	totalCutiKantor = 14 // Total cuti kantor per tahun
	maxCutiBerturut  = 3  // Max cuti berturut-turut
	daysThreshold     = 180 // Hari sebelum karyawan baru dapat mengambil cuti
)

// Function to check if an employee can take personal leave
func canTakePersonalLeave(joinDate, plannedLeaveDate time.Time, leaveDuration, cutiBersama int) (bool, string) {
	// Check if the employee is a new employee (less than 180 days)
	if plannedLeaveDate.Before(joinDate.AddDate(0, 0, daysThreshold)) {
		return false, "Karena belum 180 hari sejak tanggal join karyawan."
	}

	// Calculate the start of personal leave eligibility
	startEligibility := joinDate.AddDate(0, 0, daysThreshold)

	// Calculate the end of the year
	endOfYear := time.Date(joinDate.Year(), time.December, 31, 0, 0, 0, 0, time.Local)

	// Calculate available days for personal leave
	availableDays := int(endOfYear.Sub(startEligibility).Hours() / 24)

	// Calculate personal leave quota
	personalLeaveQuota := availableDays * cutiBersama / 365

	// Check if the requested leave duration exceeds the personal leave quota
	if leaveDuration > personalLeaveQuota {
		return false, fmt.Sprintf("Karena hanya boleh mengambil %d hari cuti", personalLeaveQuota)
	}

	// Check if the leave duration exceeds the maximum consecutive leave
	if leaveDuration > maxCutiBerturut {
		return false, fmt.Sprintf("Cuti pribadi tidak boleh lebih dari %d hari berturut-turut.", maxCutiBerturut)
	}

	// If all conditions are met, the employee can take the leave
	return true, ""
}

func validateDate(dateStr string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("format tanggal tidak valid, harap gunakan YYYY-MM-DD")
	}
	return date, nil
}

// Function to validate if input is an integer
func validateInt(input string) (int, error) {
	value, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("input harus berupa angka bulat")
	}
	return value, nil
}

func validateLeaveDuration(leaveDuration, cutiBersama int) error {
	if leaveDuration > cutiBersama {
		return fmt.Errorf("durasi cuti tidak boleh lebih dari jumlah cuti bersama: %d hari", cutiBersama)
	}
	return nil
}

func main() {
	var joinDateStr, plannedLeaveStr string
	var cutiBersamaStr, leaveDurationStr string

	for {
		// Input Jumlah Cuti Bersama
		fmt.Print("Masukkan Jumlah Cuti Bersama: ")
		fmt.Scan(&cutiBersamaStr)

		cutiBersama, err := validateInt(cutiBersamaStr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Input Tanggal Bergabung Karyawan
		var joinDate time.Time
		for {
			fmt.Print("Masukkan Tanggal Bergabung Karyawan (YYYY-MM-DD): ")
			fmt.Scan(&joinDateStr)

			var err error
			joinDate, err = validateDate(joinDateStr)
			if err != nil {
				fmt.Println(err)
				continue
			}
			break // exit inner loop if valid
		}

		// Input Tanggal Rencana Cuti
		var plannedLeaveDate time.Time
		for {
			fmt.Print("Masukkan Tanggal Rencana Cuti (YYYY-MM-DD): ")
			fmt.Scan(&plannedLeaveStr)

			var err error
			plannedLeaveDate, err = validateDate(plannedLeaveStr)
			if err != nil {
				fmt.Println(err)
				continue
			}
			break // exit inner loop if valid
		}

		// Input Durasi Cuti
		for {
			fmt.Print("Masukkan Durasi Cuti (hari): ")
			fmt.Scan(&leaveDurationStr)

			leaveDuration, err := validateInt(leaveDurationStr)
			if err != nil {
				fmt.Println(err)
				continue
			}

            // Validate leave duration against cuti bersama
			if err := validateLeaveDuration(leaveDuration, cutiBersama); err != nil {
				fmt.Println(err)
				continue
			}

			// Check if the employee can take leave
			canTakeLeave, reason := canTakePersonalLeave(joinDate, plannedLeaveDate, leaveDuration, cutiBersama)

			// Output result
            fmt.Println("=====================")
			fmt.Printf("Dapat mengambil cuti: %t\n", canTakeLeave)
            if !canTakeLeave {
                fmt.Printf("Alasan: %s\n", reason)
            }
			break // exit inner loop after processing the leave
		}
		break // exit main loop after processing all inputs
	}
}

package utils

import (
	"fmt"
	"strconv"
)

// ConvertMinutesToDuration convert minutes to duration format string
// Example: 125 -> "2h 5m"
func ConvertMinutesToDuration(durationInMinutes int64) string {

	h := durationInMinutes / 60
	m := durationInMinutes % 60

	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}

	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}

	return fmt.Sprintf("%dh %dm", h, m)
}

// ConvertHourToDuration convert hour to duration format string
// Example: 2.5 -> "2h 30m"
func ConvertHourToDuration(durationInHours float64) string {
	return ConvertMinutesToDuration(int64(durationInHours * 60))
}

// ConvertDurationToMinutes convert duration format string to minutes
// Example: "2h 30m" -> 150
func ConvertDurationToMinutes(duration string) int64 {
	var h, m int64
	fmt.Sscanf(duration, "%dh %dm", &h, &m)

	return h*60 + m
}

func FormatRupiah(amount int64) string {
	if amount == 0 {
		return "Rp0"
	}

	negative := amount < 0
	if negative {
		amount = -amount
	}

	var result []byte
	str := strconv.FormatInt(amount, 10)

	count := 0
	for i := len(str) - 1; i >= 0; i-- {
		result = append([]byte{str[i]}, result...)
		count++
		if count%3 == 0 && i != 0 {
			result = append([]byte{'.'}, result...)
		}
	}

	if negative {
		return "Rp-" + string(result)
	}
	return "Rp" + string(result)
}

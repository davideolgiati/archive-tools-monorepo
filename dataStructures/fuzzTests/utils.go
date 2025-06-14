package fuzztests

import (
	"math/rand"
	"time"
)

func getRandomString() *string {
	rand.Seed(time.Now().UnixNano())
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	length := rand.Intn(200) + 1 // Random length between 1 and 20
	result := make([]byte, length)

	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	data := string(result)

	return &data
}

package main

import (
	"math/rand"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

type number interface {
	integer | float32 | float64
}

func mod[T integer](a, b T) T {
	return (a%b + b) % b
}

func min[T number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func ternary[T any](test bool, a T, b T) T {
	if test {
		return a
	}
	return b
}

var priceFormatter = message.NewPrinter(language.Indonesian)

func formatPrice(v int64) string {
	return priceFormatter.Sprintf("Rp%d", v/100000)
}

func randstr(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}

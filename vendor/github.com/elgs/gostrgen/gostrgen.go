package gostrgen

import (
	"math/rand"
	"strings"
	"time"
)

const None = 0
const Lower = 1 << 0
const Upper = 1 << 1
const Digit = 1 << 2
const Punct = 1 << 3

const LowerUpper = Lower | Upper
const LowerDigit = Lower | Digit
const UpperDigit = Upper | Digit
const LowerUpperDigit = LowerUpper | Digit
const All = LowerUpperDigit | Punct

const lower = "abcdefghijklmnopqrstuvwxyz"
const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digit = "0123456789"
const punct = "~!@#$%^&*()_+-="

func RandGen(size int, set int, include string, exclude string) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	all := include
	if set&Lower > 0 {
		all += lower
	}
	if set&Upper > 0 {
		all += upper
	}
	if set&Digit > 0 {
		all += digit
	}
	if set&Punct > 0 {
		all += punct
	}

	lenAll := len(all)
	if len(exclude) >= lenAll {
		return ""
	}
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		b := all[rng.Intn(lenAll)]
		if strings.Contains(exclude, string(b)) {
			i--
			continue
		}
		buf[i] = b
	}
	return string(buf)
}

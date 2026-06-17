// Package unpack is a package for work with strings
package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrInvalidString is error, appearing during unpacking string.
var ErrInvalidString = errors.New("invalid string")

// Unpack is simple string unpack function. Example: "a4bc2d5e" => "aaaabccddddde".
func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	numbersOnly, _ := regexp.MatchString("^\\d*$", s)
	if numbersOnly {
		return "", ErrInvalidString
	}
	lastWasNum := false
	var builder strings.Builder
	builder.Grow(len(s))
	var prev rune
	for _, r := range s {
		if r >= 48 && r <= 57 {
			converted, err := strconv.Atoi(string(r))
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrInvalidString, err)
			}
			if lastWasNum {
				return "", fmt.Errorf("%w: нельзя использовать числа, только цифры", ErrInvalidString)
			}

			if prev == 0 {
				return "", ErrInvalidString
			}

			lastWasNum = true
			// corner case: когда встерчаем 0 то мы должны один символ удалить
			if converted == 0 {
				tempStringForDeleteLastRune := builder.String()
				_, lastRunSize := utf8.DecodeLastRuneInString(tempStringForDeleteLastRune)
				targetLen := builder.Len() - lastRunSize
				builder.Reset()
				builder.WriteString(tempStringForDeleteLastRune[:targetLen])
				continue
			}
			repeat := strings.Repeat(string(prev), converted-1)
			builder.WriteString(repeat)
			continue
		}
		lastWasNum = false
		builder.WriteRune(r)
		prev = r
	}
	return builder.String(), nil
}

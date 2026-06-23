// Package hw03frequencyanalysis need to work with most common words in string
package hw03frequencyanalysis

import (
	"fmt"
	"slices"
	"strings"
)

// Top10 takes top 10 words from string and create from it slice with sort from most common to most rare.
func Top10(in string) []string {
	// Place your code here.
	wcount := fillMap(strings.Fields(in))
	for k, v := range wcount {
		fmt.Printf("word {%s} meet {%d} times\n", k, v)
	}
	unique := make([]string, 0, len(wcount))
	for k := range wcount {
		unique = append(unique, k)
	}
	slices.SortFunc(unique, func(a string, b string) int {
		if wcount[a] != wcount[b] {
			// must retun negative number if A should be earlier then B
			// and must return positive number if A should be after B (because largest element is in the end)
			// we putting the MOST COMMON IN THE BEGINNING
			return wcount[b] - wcount[a]
		}

		return strings.Compare(a, b)
	})
	if len(unique) > 10 {
		return unique[:10]
	}
	return unique
}

func fillMap(words []string) map[string]int {
	res := make(map[string]int, len(words))
	for _, w := range words {
		res[w]++
	}
	return res
}

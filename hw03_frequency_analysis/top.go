package hw03frequencyanalysis

import (
	"strings"
)

const (
	TopLimit = 10
)

func Top10(input string) []string {
	if input == "" {
		return []string{}
	}

	words := strings.Fields(input)
	wordCount := make(map[string]int, len(words))
	for _, word := range words {
		wordCount[word]++
	}

	topWords := make([]string, 0, TopLimit)
	for i := 0; i < TopLimit; i++ {
		var bestWord string
		var bestCount int
		for word, count := range wordCount {
			if count > bestCount || (count == bestCount && word < bestWord) {
				bestWord, bestCount = word, count
			}
		}

		if bestWord != "" {
			topWords = append(topWords, bestWord)
		}
		delete(wordCount, bestWord)
	}

	if len(topWords) < TopLimit {
		return []string{}
	}

	return topWords
}

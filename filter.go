package main

import "strings"

func cleanedBody(s string) string {

	// wordsToBeRemoved := []string{"fornax","sharbert","kerfuffle"}
	var cleanedWords []string
	words := strings.Fields(s)
	for _, word := range words {
		l := strings.ToLower(word)
		if l == "fornax" || l == "sharbert" || l == "kerfuffle" {
			cleanedWords = append(cleanedWords, "****")
			continue
		}
		cleanedWords = append(cleanedWords, word)
	}

	res := strings.Join(cleanedWords, " ")
	return res
}

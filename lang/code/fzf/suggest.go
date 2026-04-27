package fzf

import "sort"

type Suggestion struct {
	Name  string
	Score float64
}

func Rank(input string, candidates []string) []Suggestion {
	var result []Suggestion
	for _, c := range candidates {
		s := Score(input, c)
		if s > 0.2 {
			result = append(result, Suggestion{Name: c, Score: s})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	return result
}

func Top(input string, candidates []string, n int) []string {
	ranked := Rank(input, candidates)
	if len(ranked) > n {
		ranked = ranked[:n]
	}
	result := make([]string, len(ranked))
	for i, r := range ranked {
		result[i] = r.Name
	}
	return result
}

package fzf

import "strings"

func TokenOverlap(a, b []string) float64 {
	setA := make(map[string]struct{})
	for _, t := range a {
		setA[Canonical(t)] = struct{}{}
	}
	setB := make(map[string]struct{})
	for _, t := range b {
		setB[Canonical(t)] = struct{}{}
	}

	intersection := 0
	for t := range setA {
		if _, ok := setB[t]; ok {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func FZFScore(input, candidate string) float64 {
	inputLower := strings.ToLower(input)
	candidateLower := strings.ToLower(candidate)

	if inputLower == "" {
		return 1.0
	}
	if len(inputLower) > len(candidateLower) {
		return 0
	}

	var positions []int
	j := 0
	for i := 0; i < len(candidateLower) && j < len(inputLower); i++ {
		if candidateLower[i] == inputLower[j] {
			positions = append(positions, i)
			j++
		}
	}
	if len(positions) != len(inputLower) {
		return 0
	}

	score := 0.0
	for idx, pos := range positions {
		score += 1.0

		if pos == 0 && idx == 0 {
			score += 3.0
		}

		if pos > 0 && isBoundary(candidate[pos-1], candidate[pos]) {
			score += 2.0
		}

		if idx > 0 && pos == positions[idx-1]+1 {
			score += 2.0
		}
	}

	score -= float64(len(candidateLower)-len(inputLower)) * 0.1

	maxPossible := float64(len(inputLower)) * 7.0
	normalized := score / maxPossible
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}
	return normalized
}

func isBoundary(prev, curr byte) bool {
	return prev == '_' || prev == '-' || prev == ' ' ||
		(prev >= 'a' && prev <= 'z' && curr >= 'A' && curr <= 'Z')
}

func Score(input, candidate string) float64 {
	inputTokens := SplitCamel(input)
	candTokens := SplitCamel(candidate)
	overlap := TokenOverlap(inputTokens, candTokens)
	fzf := FZFScore(input, candidate)
	return 0.6*overlap + 0.4*fzf
}

package util

import "slices"

func ConsecutiveWins(arr []bool) int {
	consecutives := []int{}
	current := 0 
	for i, val := range arr {
		if !val {
			consecutives = append(consecutives, current)
			current = 0
			continue
		} else {
			current += 1
		}
		if i == len(arr) - 1 {
			consecutives = append(consecutives, current)
		}
	}
	if (len(consecutives) == 0) {
		return 0
	} else {
		return slices.Max(consecutives)
	}
}

func RecentConsecutiveWins(arr []bool) int {
	score := 0 

	for _, val := range arr {
		if val {
			score += 1 
		} else {
			break
		}
	}
	return score
}

func consecutive(segment []bool) int {
	maxLen := 0
	currentLen := 0 

	for _, val := range segment {
		if val {
			currentLen ++ 
			if currentLen > maxLen {
				maxLen = currentLen 
			}
		} else {
			currentLen = 0
		}
	}
	return maxLen
}


func NearXWinX(original []bool) (near int, win int) {
	// truncate to at most 15 
	var arr []bool
	if len(original) >= 15 {
		arr = original[:15]
	} else {
		arr = original
	}

	highestAccuracy := 0.0
	for i := range arr {
		if len(arr) > 5 && i >= len(arr) - 5 {
			continue
		}

		target := arr[i:]
		length := len(target)

		wins := numWins(target)
		accuracy := float64(wins)/float64(length)

		if (accuracy > highestAccuracy) {
			highestAccuracy = accuracy
			near = length
			win = wins 
		}
	}

	return 
}

func numWins(arr []bool) (wins int) {
	for _, elem := range arr {
		if elem {
			wins += 1
		}
	}
	return
}

func Accuracy(original []bool) (accuracy int) {
	// truncate to latest 10 
	var arr []bool
	if (len(original) > 10) {
		arr = original[:10]
	} else {
		arr = original
	}

	wins := 0 
	for _, val := range arr {
		if (val) {
			wins += 1
		}
	}
	return int(float64(wins)/float64(len(arr)) * 100)
}
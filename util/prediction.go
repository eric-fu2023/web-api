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


func NearXWinX(arr []bool) (near int, win int) {
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


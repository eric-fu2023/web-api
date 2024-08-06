package util

func ConsecutiveWins(arr []bool) int {
	return 1
}

func NearXWinX(arr []bool) (near int, win int) {
	highestAccuracy := 0.0
	for i := range arr {
		if i >= len(arr) - 5 {
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


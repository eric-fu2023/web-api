package util

func ConsecutiveWins(arr []bool) int {
	loseIndices := []int{}
	for i, val := range arr {
		if !val {
			loseIndices = append(loseIndices, i)
		}
	}

	var firstSeg, secondSeg []bool 

	if len(loseIndices) == 0{
		firstSeg = arr
		secondSeg = []bool{}
	} else {
		firstLose := loseIndices[0]
		secondLose := len(arr)
		if len(loseIndices) > 1 {
			secondLose = loseIndices[1]
		}
		firstSeg = arr[:firstLose]
		secondSeg = arr[firstLose+1 : secondLose]
	}

	return Max(consecutive(firstSeg), consecutive(secondSeg))
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


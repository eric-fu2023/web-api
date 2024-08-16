package util_test

import (
	"fmt"
	"testing"
	"web-api/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type consecutiveWinCase struct {
	data     []bool
	expected int
}

var consecutiveWinCases = []consecutiveWinCase{
	{[]bool{true, true, true, true, true, false, true, true, true, false}, 5},
	{[]bool{true, true, false, true, true, true, true, true, true, true, false}, 7},
	{[]bool{true, true, false, false, true, true, true, true, true}, 5},
	{[]bool{true, true, true, true, true, true, true, true, true}, 9},
	{[]bool{false, false, false, false, false, false, false, false, false}, 0},
	{[]bool{false, false, true, true, false, false, false, false, false}, 2},
	{[]bool{true, true, true, false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false}, 19},
	{[]bool{true, true, true, false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}, 28},
	{[]bool{true, true, false, true, true, false, true, false, true, false, false, false, false}, 2},
	{[]bool{true, false, true, true, false, true, true, true, true, true, true, true}, 7},
	{[]bool{false, false, true, true, true, true, true, true, true}, 7},
	{[]bool{}, 0},
}

func TestConsecutiveWins(t *testing.T) {
	fmt.Println("=======TestConsecutiveWins=======")
	for i, test := range consecutiveWinCases {
		t.Run(fmt.Sprintf("ConsecutiveWins-Case %d", i), func(t *testing.T) {
			streak := util.ConsecutiveWins(test.data)
			require.NoError(t, nil)
			assert.Equal(t, test.expected, streak)

		})
	}
}

type nearXWinXCase struct {
	data         []bool
	expectedNear int
	expectedWin  int
}

var nearXWinXCases = []nearXWinXCase{
	{[]bool{true, true, true, false, true, true, false, true, true, false, false}, 11, 7},
	{[]bool{true, false, true, true, true, false, true, true, false, false, true, false}, 10, 6},
	{[]bool{true, false, true, true, true, true, true, true, true, true, false, true}, 10, 9},
	{[]bool{true, true, true, true, true, true, true, true, true, true, true, true}, 12, 12},
	{[]bool{false, false, false, false, false, false, false, false, false, false, true, true}, 6, 2},
	{[]bool{true, true, false, false, false, true, false, true, false, false, true, true}, 7, 4},
	{[]bool{}, 0, 0},
	{[]bool{true}, 1, 1},
	{[]bool{false, true}, 1, 1},
	{[]bool{true, false}, 2, 1},
	{[]bool{true, false, true}, 1, 1},
}

func TestNearXWinX(t *testing.T) {
	fmt.Println("=======TestNearXWinX=======")

	for i, test := range nearXWinXCases {
		t.Run(fmt.Sprintf("NearXWinX-Case%d", i), func(t *testing.T) {
			near, win := util.NearXWinX(test.data)
			require.NoError(t, nil)
			assert.Equal(t, test.expectedNear, near)
			assert.Equal(t, test.expectedWin, win)
		})
	}
}

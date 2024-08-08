package util_test

import (
	"fmt"
	"testing"
	"web-api/util"
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
	{[]bool{true,true,true,false,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,false}, 19},
	{[]bool{true,true,true,false,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,}, 28},
	{[]bool{true,true,false,true,true,false,true,false,true,false,false,false,false,}, 2},
	{[]bool{true,false,true,true,false,true,true,true,true,true,true,true}, 7}, 
	{[]bool{false, false, true, true, true, true, true, true, true}, 7},
}

func TestConsecutiveWins(t *testing.T) {
	fmt.Println("=======TestConsecutiveWins=======")
	for i, test := range consecutiveWinCases {
		fmt.Printf("Testing case %v : \n", test.data)
		if output := util.ConsecutiveWins(test.data); output != test.expected {
			fmt.Printf("ERROR!! Expected %d, got %d\n", test.expected, output)
			t.Errorf("Test %d - Output %d not equal to expected %d", i, output, test.expected)
		} else {
			fmt.Printf("OK\n")
		}
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
}

func TestNearXWinX(t *testing.T) {
	fmt.Println("=======TestNearXWinX=======")

	for i, test := range nearXWinXCases {
		fmt.Printf("Testing case %v : \n", test.data)
		if outNear, outWin := util.NearXWinX(test.data); outNear != test.expectedNear || outWin != test.expectedWin {
			fmt.Printf("ERROR!! Expected (%d,%d), got (%d,%d)\n", test.expectedNear, test.expectedWin, outNear, outWin)
			t.Errorf("Test %d - Output (%d,%d) not equal to expected (%d,%d)", i, outNear, outWin, test.expectedNear, test.expectedWin)
		} else {
			fmt.Printf("OK\n")
		}
	}
}

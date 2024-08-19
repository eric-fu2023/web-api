package test

import (
	"testing"
	"web-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProgressCase struct {
	name            string
	currentProgress int64
	expected        []int64
}

var fakeProgressCases = []fakeProgressCase{
	{"0 Progress (Initial)", 0, []int64{0, 93}},
	{"Test Fake Progress 1", 9999, []int64{9999, 9999}},
	{"Test Fake Progress 2", 9989, []int64{9989, 9999}},
	{"Test Fake Progress 3", 9989, []int64{9989, 9990}},
	{"Test Fake Progress 4", 8954, []int64{8954, 8955}},
	{"Test Fake Progress 5", 8954, []int64{8954, 9054}},
}

func TestGenerateFakeProgress(t *testing.T) {

	initialLowerLimit := model.InitialRandomFakeProgressLowerLimit
	initialUpperLimit := model.InitialRandomFakeProgressUpperLimit

	subsequentLowerLimit := model.SubsequentRandomFakeProgressLowerLimit
	subsequentUpperLimit := model.SubsequentRandomFakeProgressUpperLimit

	for _, tc := range fakeProgressCases {
		t.Run(tc.name, func(t *testing.T) {

			beforeProgress, afterProgress := model.GenerateFakeProgress(tc.currentProgress)
			actual := afterProgress

			t.Logf("Before Progress - %v \n", beforeProgress)
			t.Logf("After Progress - %v \n", afterProgress)

			require.NoError(t, nil)
			assert.Equal(t, beforeProgress, tc.expected[0])

			if tc.currentProgress == 0 {
				assert.LessOrEqual(t, initialLowerLimit, actual, "afterProgress is greater than the initial lower limit")
				assert.GreaterOrEqual(t, initialUpperLimit, actual, "afterProgress is lesser than the initial upper limit")
			} else if tc.currentProgress == 9999 {
				assert.Equal(t, tc.expected[0], actual)
			} else {
				assert.LessOrEqual(t, beforeProgress+subsequentLowerLimit, actual, "afterProgress is greater than the subsequent lower limit")
				assert.GreaterOrEqual(t, beforeProgress+subsequentUpperLimit, actual, "afterProgress is lesser than the subsequent upper limit")
			}
		})
	}
}

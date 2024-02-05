package cashout_test

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestWithdraw(t *testing.T) {
	a := json.RawMessage(`{"a":"b"}`)
	c := struct {
		a json.RawMessage
		b string
	}{
		a: a,
		b: "xxx",
	}
	t.Log(fmt.Sprintf("%s", c))
	t.Error(1)
}

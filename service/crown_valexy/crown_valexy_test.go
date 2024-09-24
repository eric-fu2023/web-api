package crown_valexy

import (
	"encoding/json"
	"testing"
)

func TestRemarks_String(t *testing.T) {
	type fields struct {
		WithdrawId string
		DepositId  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "1",
			fields: fields{
				WithdrawId: "w1",
				DepositId:  "w1",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Remarks{
				WithdrawId: tt.fields.WithdrawId,
				DepositId:  tt.fields.DepositId,
			}

			_got, _ := json.Marshal(r)
			got := string(_got)
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

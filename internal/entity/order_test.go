package entity

import "testing"

func TestValidateLuhn(t *testing.T) {
	type args struct {
		number string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "valid 79927398713", args: args{"79927398713"}, want: true},
		{name: "valid 12345678903", args: args{"12345678903"}, want: true},
		{name: "valid 9278923470", args: args{"9278923470"}, want: true},
		{name: "invalid 12345", args: args{"12345"}, want: false},
		{name: "invalid 111", args: args{"111"}, want: false},
		{name: "empty string", args: args{""}, want: false},
		{name: "letters", args: args{"abc"}, want: false},
		{name: "mixed", args: args{"1234a5678"}, want: false},
		{name: "single zero", args: args{"0"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateLuhn(tt.args.number); got != tt.want {
				t.Errorf("ValidateLuhn() = %v, want %v", got, tt.want)
			}
		})
	}
}

package service

import "testing"

func TestParseCheckboxValue(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want float64
	}{
		{"bool true", true, 1},
		{"bool false", false, 0},
		{"float 1", float64(1), 1},
		{"float 0", float64(0), 0},
		{"str true", "true", 1},
		{"str 1", "1", 1},
		{"str other", "no", 0},
		{"nil", nil, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parseCheckboxValue(c.in)
			if got != c.want {
				t.Errorf("parseCheckboxValue(%#v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

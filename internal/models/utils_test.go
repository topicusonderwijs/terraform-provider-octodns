package models

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestRefFloat64(t *testing.T) {
	want := 10.1
	got := RefFloat64(want)
	if *got != want {
		t.Errorf("RefFloat64: unexpected value got %v, want %v", got, want)
	}
}
func TestRefInt(t *testing.T) {
	want := 10
	got := RefInt(want)
	if *got != want {
		t.Errorf("RefInt: unexpected value got %v, want %v", got, want)
	}
}
func TestRefString(t *testing.T) {
	want := "example"
	got := RefString(want)
	if *got != want {
		t.Errorf("RefString: unexpected value got %v, want %v", got, want)
	}
}

func TestRefStringAsFloat(t *testing.T) {

	tests := []struct {
		Value   string
		WantNil bool
		Want    float64
	}{
		{
			Value:   "10.1",
			WantNil: false,
			Want:    10.1,
		},
		{
			Value:   "1337",
			WantNil: false,
			Want:    1337.0,
		},
		{
			Value:   "-20",
			WantNil: false,
			Want:    -20.0,
		},
		{
			Value:   "example",
			WantNil: true,
		},
		{
			Value:   "2nd-example",
			WantNil: true,
		},
		{
			Value:   "",
			WantNil: true,
		},
	}

	for _, test := range tests {

		got := RefStringAsFloat64(test.Value)

		if test.WantNil {
			if got != nil {
				t.Errorf("RefStringAsFloat64: Expecting nil value but got %v", *got)
			}
		} else {

			if *got != test.Want {
				t.Errorf("RefStringAsFloat64: unexpected value got %v, want %v", *got, test.Value)
			}
		}

	}

}

func TestRefStringAsInt(t *testing.T) {

	tests := []struct {
		Value   string
		WantNil bool
		Want    int
	}{
		{
			Value:   "10",
			WantNil: false,
			Want:    10,
		},
		{
			Value:   "1337",
			WantNil: false,
			Want:    1337,
		},
		{
			Value:   "-20",
			WantNil: false,
			Want:    -20,
		},
		{
			Value:   "example",
			WantNil: true,
		},
		{
			Value:   "2nd-example",
			WantNil: true,
		},
		{
			Value:   "10.5",
			WantNil: true,
		},
		{
			Value:   "",
			WantNil: true,
		},
	}

	for _, test := range tests {

		got := RefStringAsInt(test.Value)

		if test.WantNil {
			if got != nil {
				t.Errorf("RefStringAsInt: Expecting nil value but got %v", *got)
			}
		} else {
			if got == nil {
				t.Errorf("RefStringAsInt: Unexpected nil value for %v", test.Value)
			} else if *got != test.Want {
				t.Errorf("RefStringAsInt: unexpected value got %v, want %v", *got, test.Value)
			}
		}

	}

}

func TestRegexToMap(t *testing.T) {

	tests := []struct {
		Value   string
		Pattern string
		WantErr bool
		Want    map[string]string
	}{
		{
			Value:   "10 example.com.",
			Pattern: "^(?P<preference>\\d+) (?P<exchange>.+)$",
			WantErr: false,
			Want: map[string]string{
				"preference": "10",
				"exchange":   "example.com.",
			},
		},
		{
			Value:   "10 10 example.com.",
			Pattern: "^(?P<preference>\\d+) (?P<exchange>.[^ ]+)$",
			WantErr: true,
			Want: map[string]string{
				"preference": "10",
				"exchange":   "example.com.",
			},
		},
	}

	for _, test := range tests {

		got, err := regexToMap(test.Value, test.Pattern)

		if test.WantErr {
			if err == nil {
				t.Errorf("regexToMap: Expecting error but got %v", got)
			}

		} else {
			if err != nil {
				t.Errorf("regexToMap: Unexpected error for test: %s", err.Error())
			} else {
				diff := cmp.Diff(got, test.Want)
				if diff != "" {
					t.Errorf("regexToMap: difference found: %s", diff)
				}
			}
		}

	}

}

package hw09structvalidator

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateValidStructure(t *testing.T) {
	t.Parallel()
	type request struct {
		Name   string   `validate:"len:5|in:Alice,Bobby"`
		Age    int      `validate:"min:18|max:50|in:20,30"`
		Codes  []int    `validate:"min:100|max:599"`
		Phones []string `validate:"len:3|regexp:^\\d+$"`
		Note   string
	}
	value := request{"Alice", 20, []int{200, 404}, []string{"123", "456"}, "ignored"}
	if err := Validate(value); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAccumulatesErrors(t *testing.T) {
	t.Parallel()
	type request struct {
		Name string `validate:"len:5|regexp:^\\d+$"`
		Age  int    `validate:"min:18|max:50|in:20,30"`
	}
	err := Validate(request{Name: "ab", Age: 60})
	var validationErrors ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}
	if len(validationErrors) != 4 {
		t.Fatalf("expected 4 errors, got %d: %v", len(validationErrors), err)
	}
	for _, target := range []error{ErrLen, ErrRegexp, ErrMax, ErrIntIn} {
		if !errors.Is(err, target) {
			t.Errorf("errors.Is(err, %v) is false", target)
		}
	}
}

func TestValidateSlices(t *testing.T) {
	t.Parallel()
	type request struct {
		Numbers []int    `validate:"min:10"`
		Words   []string `validate:"in:foo,bar"`
	}
	err := Validate(request{[]int{1, 20, 2}, []string{"bad", "foo", "no"}})
	var validationErrors ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) != 4 {
		t.Fatalf("expected 4 ValidationErrors, got %T: %v", err, err)
	}
	for _, index := range []string{"element 0", "element 2"} {
		if !strings.Contains(err.Error(), index) {
			t.Errorf("error does not contain %q: %v", index, err)
		}
	}
}

func TestValidateIgnoresFields(t *testing.T) {
	t.Parallel()
	type request struct {
		WithoutTag string
		private    int `validate:"min:100"`
	}
	if err := Validate(request{private: 1}); err != nil {
		t.Fatalf("ignored fields produced an error: %v", err)
	}
}

func TestValidateProgramErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		value  interface{}
		target error
	}{
		{"nil", nil, ErrNotStruct},
		{"not struct", 42, ErrNotStruct},
		{"pointer", &struct{}{}, ErrNotStruct},
		{"missing colon", struct {
			Age int `validate:"min=18"`
		}{20}, ErrInvalidRule},
		{"unknown rule", struct {
			Age int `validate:"positive:1"`
		}{20}, ErrInvalidRule},
		{"bad integer", struct {
			Age int `validate:"min:adult"`
		}{20}, ErrInvalidRuleArgument},
		{"bad regexp", struct {
			Text string `validate:"regexp:["`
		}{"x"}, ErrInvalidRuleArgument},
		{"unsupported field", struct {
			Enabled bool `validate:"in:true"`
		}{true}, ErrUnsupportedType},
		{"unsupported empty slice", struct {
			Bytes []byte `validate:"min:1"`
		}{}, ErrUnsupportedType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.value)
			if !errors.Is(err, tt.target) {
				t.Fatalf("expected %v, got %T: %v", tt.target, err, err)
			}
			var validationErrors ValidationErrors
			if errors.As(err, &validationErrors) {
				t.Fatalf("program error must not be ValidationErrors: %v", err)
			}
		})
	}
}

func TestValidationRuleErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		value  interface{}
		target error
	}{
		{ruleMin, struct {
			V int `validate:"min:10"`
		}{9}, ErrMin},
		{ruleMax, struct {
			V int `validate:"max:10"`
		}{11}, ErrMax},
		{"integer in", struct {
			V int `validate:"in:10,20"`
		}{11}, ErrIntIn},
		{"unicode length", struct {
			V string `validate:"len:6"`
		}{"привет"}, nil},
		{ruleRegexp, struct {
			V string `validate:"regexp:^\\d+$"`
		}{"abc"}, ErrRegexp},
		{"string in", struct {
			V string `validate:"in:foo,bar"`
		}{"baz"}, ErrStringIn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.value)
			if tt.target == nil {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.target) {
				t.Fatalf("expected %v, got %v", tt.target, err)
			}
		})
	}
}

// Package hw09structvalidator validates structure fields using validate tags.
package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	// ErrNotStruct -> receieved not struct value.
	ErrNotStruct = errors.New("value is not a struct")
	// ErrInvalidRule -> unknow rule.
	ErrInvalidRule = errors.New("invalid validation rule")
	// ErrInvalidRuleArgument -> incorrect format.
	ErrInvalidRuleArgument = errors.New("invalid validation rule argument")
	// ErrUnsupportedType -> type not supported.
	ErrUnsupportedType = errors.New("unsupported field type")

	// ErrLen is len error.
	ErrLen = errors.New("invalid string length")
	// ErrRegexp not match regex.
	ErrRegexp = errors.New("string does not match regexp")
	// ErrStringIn not contains in set.
	ErrStringIn = errors.New("string is not in allowed set")
	// ErrMin less then minimum.
	ErrMin = errors.New("integer is less than minimum")
	// ErrMax more than max.
	ErrMax = errors.New("integer is greater than maximum")
	// ErrIntIn not contains in set.
	ErrIntIn = errors.New("integer is not in allowed set")
)

// ValidationError represent validation error.
type ValidationError struct {
	Field string
	Err   error
}

func (v ValidationError) Error() string { return fmt.Sprintf("field %s: %v", v.Field, v.Err) }
func (v ValidationError) Unwrap() error { return v.Err }

// ValidationErrors all validation errors.
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for i, validationError := range v {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(validationError.Error())
	}
	return sb.String()
}

// Unwrap for working with all slice.
func (v ValidationErrors) Unwrap() []error {
	errs := make([]error, 0, len(v))
	for _, validationError := range v {
		errs = append(errs, validationError)
	}
	return errs
}

type (
	stringRule func(string, []string) (bool, error)
	intRule    func(int, []string) (bool, error)
)

var validationRulesString = map[string]stringRule{
	"len": validateLen, "regexp": validateRegexp, "in": validateStringIn,
}

var validationRulesInt = map[string]intRule{
	"min": validateMin, "max": validateMax, "in": validateIntIn,
}

// Validate check structs with "validate" tag.
func Validate(v interface{}) error {
	if v == nil {
		return fmt.Errorf("%w: got nil", ErrNotStruct)
	}
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("%w: got %T", ErrNotStruct, v)
	}

	value := reflect.ValueOf(v)
	var result ValidationErrors
	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		fieldErrors, err := validateField(field, value.FieldByIndex(field.Index), tag)
		if err != nil {
			return err
		}
		result = append(result, fieldErrors...)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func validateField(field reflect.StructField, value reflect.Value, tag string) (ValidationErrors, error) {
	//nolint:exhaustive
	switch field.Type.Kind() {
	case reflect.String:
		return validateString(field.Name, value.String(), tag)
	case reflect.Int:
		return validateInt(field.Name, int(value.Int()), tag)
	case reflect.Slice:
		return validateSlice(field.Name, value, tag)
	default:
		return nil, fmt.Errorf("%w: field %s has type %s", ErrUnsupportedType, field.Name, field.Type)
	}
}

func validateSlice(fieldName string, value reflect.Value, tag string) (ValidationErrors, error) {
	elementKind := value.Type().Elem().Kind()
	if elementKind != reflect.String && elementKind != reflect.Int {
		return nil, fmt.Errorf("%w: field %s has element type %s", ErrUnsupportedType, fieldName, value.Type().Elem())
	}

	var result ValidationErrors
	for i := 0; i < value.Len(); i++ {
		element := value.Index(i)
		var elementErrors ValidationErrors
		var err error
		//nolint:exhaustive
		switch element.Kind() {
		case reflect.String:
			elementErrors, err = validateString(fieldName, element.String(), tag)
		case reflect.Int:
			elementErrors, err = validateInt(fieldName, int(element.Int()), tag)
		default:
			return nil, fmt.Errorf("%w: field %s has element type %s", ErrUnsupportedType, fieldName, element.Type())
		}
		if err != nil {
			return nil, err
		}
		for _, elementError := range elementErrors {
			elementError.Err = fmt.Errorf("element %d: %w", i, elementError.Err)
			result = append(result, elementError)
		}
	}
	return result, nil
}

func validateInt(fieldName string, value int, tag string) (ValidationErrors, error) {
	var result ValidationErrors
	for _, rawRule := range strings.Split(tag, "|") {
		name, args, err := parseRule(rawRule)
		if err != nil {
			return nil, err
		}
		rule, ok := validationRulesInt[name]
		if !ok {
			return nil, fmt.Errorf("%w %q for field %s", ErrInvalidRule, name, fieldName)
		}
		passed, err := rule(value, args)
		if err != nil {
			return nil, fmt.Errorf("field %s, rule %s: %w", fieldName, name, err)
		}
		if !passed {
			result = append(result, ValidationError{Field: fieldName, Err: intRuleError(name, value, args)})
		}
	}
	return result, nil
}

func validateString(fieldName, value, tag string) (ValidationErrors, error) {
	var result ValidationErrors
	for _, rawRule := range strings.Split(tag, "|") {
		name, args, err := parseRule(rawRule)
		if err != nil {
			return nil, err
		}
		rule, ok := validationRulesString[name]
		if !ok {
			return nil, fmt.Errorf("%w %q for field %s", ErrInvalidRule, name, fieldName)
		}
		passed, err := rule(value, args)
		if err != nil {
			return nil, fmt.Errorf("field %s, rule %s: %w", fieldName, name, err)
		}
		if !passed {
			result = append(result, ValidationError{Field: fieldName, Err: stringRuleError(name, value, args)})
		}
	}
	return result, nil
}

func parseRule(rawRule string) (string, []string, error) {
	parts := strings.SplitN(rawRule, ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", nil, fmt.Errorf("%w: %q", ErrInvalidRule, rawRule)
	}
	return parts[0], strings.Split(parts[1], ","), nil
}

func validateLen(value string, args []string) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("%w: len expects one argument", ErrInvalidRuleArgument)
	}
	want, err := strconv.Atoi(args[0])
	if err != nil || want < 0 {
		return false, fmt.Errorf("%w: invalid len argument %q", ErrInvalidRuleArgument, args[0])
	}
	return utf8.RuneCountInString(value) == want, nil
}

func validateRegexp(value string, args []string) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("%w: regexp expects one argument", ErrInvalidRuleArgument)
	}
	re, err := regexp.Compile(args[0])
	if err != nil {
		return false, fmt.Errorf("%w: invalid regexp %q: %w", ErrInvalidRuleArgument, args[0], err)
	}
	return re.MatchString(value), nil
}

func validateStringIn(value string, args []string) (bool, error) {
	if len(args) == 0 || len(args) == 1 && args[0] == "" {
		return false, fmt.Errorf("%w: in expects an argument", ErrInvalidRuleArgument)
	}
	for _, allowed := range args {
		if value == allowed {
			return true, nil
		}
	}
	return false, nil
}

func validateMin(value int, args []string) (bool, error) {
	limit, err := parseSingleIntArgument("min", args)
	return value >= limit, err
}

func validateMax(value int, args []string) (bool, error) {
	limit, err := parseSingleIntArgument("max", args)
	return value <= limit, err
}

func validateIntIn(value int, args []string) (bool, error) {
	if len(args) == 0 || len(args) == 1 && args[0] == "" {
		return false, fmt.Errorf("%w: in expects an argument", ErrInvalidRuleArgument)
	}
	for _, arg := range args {
		allowed, err := strconv.Atoi(arg)
		if err != nil {
			return false, fmt.Errorf("%w: invalid in argument %q", ErrInvalidRuleArgument, arg)
		}
		if value == allowed {
			return true, nil
		}
	}
	return false, nil
}

func parseSingleIntArgument(name string, args []string) (int, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("%w: %s expects one argument", ErrInvalidRuleArgument, name)
	}
	value, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s argument %q", ErrInvalidRuleArgument, name, args[0])
	}
	return value, nil
}

func stringRuleError(name, value string, args []string) error {
	switch name {
	case "len":
		return fmt.Errorf("%w: %q must contain %s characters", ErrLen, value, args[0])
	case "regexp":
		return fmt.Errorf("%w: %q does not match %q", ErrRegexp, value, args[0])
	case "in":
		return fmt.Errorf("%w: %q is not one of %s", ErrStringIn, value, strings.Join(args, ","))
	default:
		panic("unreachable string rule")
	}
}

func intRuleError(name string, value int, args []string) error {
	switch name {
	case "min":
		return fmt.Errorf("%w: %d must be at least %s", ErrMin, value, args[0])
	case "max":
		return fmt.Errorf("%w: %d must be at most %s", ErrMax, value, args[0])
	case "in":
		return fmt.Errorf("%w: %d is not one of %s", ErrIntIn, value, strings.Join(args, ","))
	default:
		panic("unreachable integer rule")
	}
}

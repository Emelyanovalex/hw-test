package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotStruct    = errors.New("input is not a struct")
	ErrInvalidTag   = errors.New("invalid validate tag")
	ErrStringLen    = errors.New("string length mismatch")
	ErrStringRegexp = errors.New("string does not match regexp")
	ErrStringIn     = errors.New("string not in allowed set")
	ErrIntMin       = errors.New("int value is below minimum")
	ErrIntMax       = errors.New("int value is above maximum")
	ErrIntIn        = errors.New("int value not in allowed set")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for i, ve := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		fmt.Fprintf(&sb, "%s: %v", ve.Field, ve.Err)
	}
	return sb.String()
}

func Validate(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	rt := rv.Type()
	var validationErrors ValidationErrors

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		value := rv.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		errs, err := validateField(field.Name, value, tag)
		if err != nil {
			return err
		}
		validationErrors = append(validationErrors, errs...)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

func validateField(name string, value reflect.Value, tag string) (ValidationErrors, error) {
	switch value.Kind() { //nolint:exhaustive
	case reflect.String:
		return validateStringField(name, value.String(), tag)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateIntField(name, value.Int(), tag)
	case reflect.Slice:
		return validateSliceField(name, value, tag)
	default:
		return nil, nil
	}
}

func validateStringField(name, s, tag string) (ValidationErrors, error) {
	rules := strings.Split(tag, "|")
	var errs ValidationErrors

	for _, rule := range rules {
		parts := strings.SplitN(rule, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: %q", ErrInvalidTag, rule)
		}

		var validationErr error
		var progErr error

		switch parts[0] {
		case "len":
			validationErr, progErr = validateStringLen(s, parts[1])
		case "regexp":
			validationErr, progErr = validateStringRegexp(s, parts[1])
		case "in":
			validationErr = validateStringIn(s, parts[1])
		default:
			return nil, fmt.Errorf("%w: unknown rule %q", ErrInvalidTag, parts[0])
		}

		if progErr != nil {
			return nil, progErr
		}
		if validationErr != nil {
			errs = append(errs, ValidationError{Field: name, Err: validationErr})
		}
	}

	return errs, nil
}

func validateStringLen(s, param string) (error, error) {
	n, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf("%w: len must be an integer, got %q", ErrInvalidTag, param)
	}
	if len([]rune(s)) != n {
		return fmt.Errorf("%w: expected %d, got %d", ErrStringLen, n, len([]rune(s))), nil
	}
	return nil, nil
}

func validateStringRegexp(s, pattern string) (error, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid regexp %q: %w", ErrInvalidTag, pattern, err)
	}
	if !re.MatchString(s) {
		return fmt.Errorf("%w: %q does not match %q", ErrStringRegexp, s, pattern), nil
	}
	return nil, nil
}

func validateStringIn(s, param string) error {
	for _, v := range strings.Split(param, ",") {
		if s == v {
			return nil
		}
	}
	return fmt.Errorf("%w: %q not in [%s]", ErrStringIn, s, param)
}

func validateIntField(name string, n int64, tag string) (ValidationErrors, error) {
	rules := strings.Split(tag, "|")
	var errs ValidationErrors

	for _, rule := range rules {
		parts := strings.SplitN(rule, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: %q", ErrInvalidTag, rule)
		}

		var validationErr error
		var progErr error

		switch parts[0] {
		case "min":
			validationErr, progErr = validateIntMin(n, parts[1])
		case "max":
			validationErr, progErr = validateIntMax(n, parts[1])
		case "in":
			validationErr, progErr = validateIntIn(n, parts[1])
		default:
			return nil, fmt.Errorf("%w: unknown rule %q", ErrInvalidTag, parts[0])
		}

		if progErr != nil {
			return nil, progErr
		}
		if validationErr != nil {
			errs = append(errs, ValidationError{Field: name, Err: validationErr})
		}
	}

	return errs, nil
}

func validateIntMin(n int64, param string) (error, error) {
	minVal, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: min must be an integer, got %q", ErrInvalidTag, param)
	}
	if n < minVal {
		return fmt.Errorf("%w: %d < %d", ErrIntMin, n, minVal), nil
	}
	return nil, nil
}

func validateIntMax(n int64, param string) (error, error) {
	maxVal, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: max must be an integer, got %q", ErrInvalidTag, param)
	}
	if n > maxVal {
		return fmt.Errorf("%w: %d > %d", ErrIntMax, n, maxVal), nil
	}
	return nil, nil
}

func validateIntIn(n int64, param string) (error, error) {
	for _, v := range strings.Split(param, ",") {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: in value must be an integer, got %q", ErrInvalidTag, v)
		}
		if n == val {
			return nil, nil
		}
	}
	return fmt.Errorf("%w: %d not in [%s]", ErrIntIn, n, param), nil
}

func validateSliceField(name string, value reflect.Value, tag string) (ValidationErrors, error) {
	var errs ValidationErrors
	for i := 0; i < value.Len(); i++ {
		elemErrs, err := validateField(name, value.Index(i), tag)
		if err != nil {
			return nil, err
		}
		errs = append(errs, elemErrs...)
	}
	return errs, nil
}

package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Form represents a form with validation errors.
type Form struct {
	url.Values        // Embed url.Values to handle form data.
	Errors     errors // Custom error handling for the form.
}

func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d)", d))
	}
}

func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

// New creates a new Form instance with the provided data and initializes errors.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}), // Initialize Errors with an empty map.
	}
}

// Required checks if the specified fields are present and not blank.
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank") // Add error if the field is blank.
		}
	}
}

// MaxLength checks if the specified field's value exceeds the maximum length.
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return // Skip if the field is empty.
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d)) // Add error if the value is too long.
	}
}

// PermittedValues checks if the value of the specified field is among permitted values.
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return // Skip if the field is empty.
	}
	for _, opt := range opts {
		if value == opt {
			return // Return if the value matches one of the permitted options.
		}
	}
	f.Errors.Add(field, "This field is invalid") // Add error if the value is not permitted.
}

// Valid checks if there are no validation errors.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0 // Return true if there are no errors.
}

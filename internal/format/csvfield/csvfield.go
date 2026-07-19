package csvfield

import (
	"encoding/csv"
	"strings"
)

// Split returns the comma-separated fields of value, honouring the quoting
// rules of [encoding/csv]. An empty value yields no fields.
//
// It returns any error reported while reading value as a CSV record.
func Split(value string) ([]string, error) {
	if value == "" {
		return []string{}, nil
	}
	return csv.NewReader(strings.NewReader(value)).Read()
}

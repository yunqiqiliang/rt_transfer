package typing

import (
	"strconv"
	"strings"

	"github.com/artie-labs/transfer/lib/typing/decimal"
)

const defaultPrefix = "numeric"

// ParseNumeric - will prefix (since it can be NUMBER or NUMERIC) + valString in the form of:
// * NUMERIC(p, s)
// * NUMERIC(p)
func ParseNumeric(prefix, valString string) KindDetails {
	if !strings.HasPrefix(valString, prefix) {
		return Invalid
	}

	valString = strings.TrimPrefix(valString, prefix+"(")
	valString = strings.TrimSuffix(valString, ")")
	parts := strings.Split(valString, ",")
	if len(parts) == 0 || len(parts) > 2 {
		return Invalid
	}

	var parsedNumbers []int
	for _, part := range parts {
		parsedNumber, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return Invalid
		}

		parsedNumbers = append(parsedNumbers, parsedNumber)
	}

	// If scale is 0 or not specified, then number is an int.
	if len(parsedNumbers) == 1 || parsedNumbers[1] == 0 {
		return Integer
	}

	eDec := EDecimal
	eDec.ExtendedDecimalDetails = decimal.NewDecimal(parsedNumbers[1], &parsedNumbers[0], nil)
	return eDec
}

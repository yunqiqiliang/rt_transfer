package redshift

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/artie-labs/transfer/lib/stringutil"

	"github.com/artie-labs/transfer/lib/config/constants"
	"github.com/artie-labs/transfer/lib/typing"
	"github.com/artie-labs/transfer/lib/typing/columns"
)

func BenchmarkMethods(b *testing.B) {
	// Random string of length [500, 100,000)
	colVal := stringutil.Random(rand.Intn(100000) + 500) // use the same value for both benchmarks
	colKind := columns.NewColumn("foo", typing.String)   // use the same column kind for both benchmarks

	b.Run("OldMethod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			replaceExceededValuesOld(colVal, colKind)
		}
	})

	b.Run("NewMethod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			replaceExceededValuesNew(colVal, colKind)
		}
	})
}

func replaceExceededValuesOld(colVal interface{}, colKind columns.Column) interface{} {
	colValString := fmt.Sprint(colVal)
	switch colKind.KindDetails.Kind {
	case typing.Struct.Kind:
		if len(colValString) > maxRedshiftSuperLen {
			return map[string]interface{}{
				"key": constants.ExceededValueMarker,
			}
		}
	case typing.String.Kind:
		if len(colValString) > maxRedshiftVarCharLen {
			return constants.ExceededValueMarker
		}
	}

	return colVal
}

func replaceExceededValuesNew(colVal interface{}, colKind columns.Column) interface{} {
	colValString := fmt.Sprint(colVal)
	colValBytes := len(colValString)
	switch colKind.KindDetails.Kind {
	case typing.Struct.Kind:
		if colValBytes > maxRedshiftSuperLen {
			return map[string]interface{}{
				"key": constants.ExceededValueMarker,
			}
		}
	case typing.String.Kind:
		if colValBytes > maxRedshiftVarCharLen {
			return constants.ExceededValueMarker
		}
	}

	return colVal
}

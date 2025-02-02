package ddl_test

import (
	"fmt"

	"github.com/artie-labs/transfer/lib/destination/ddl"
	"github.com/stretchr/testify/assert"

	"github.com/artie-labs/transfer/lib/config/constants"
)

func (d *DDLTestSuite) TestColumnAlreadyExistErr() {
	type _testCase struct {
		name           string
		err            error
		kind           constants.DestinationKind
		expectedResult bool
	}

	testCases := []_testCase{
		{
			name:           "Redshift actual error",
			err:            fmt.Errorf(`ERROR: column "foo" of relation "statement" already exists [ErrorId: 1-64da9ea9]`),
			kind:           constants.Redshift,
			expectedResult: true,
		},
		{
			name: "Redshift error, but irrelevant",
			err:  fmt.Errorf("foo"),
			kind: constants.Redshift,
		},
	}

	for _, tc := range testCases {
		actual := ddl.ColumnAlreadyExistErr(tc.err, tc.kind)
		assert.Equal(d.T(), tc.expectedResult, actual, tc.name)
	}
}

package columns

import (
	"github.com/artie-labs/transfer/lib/config/constants"

	"github.com/artie-labs/transfer/lib/typing"

	"github.com/artie-labs/transfer/lib/typing/ext"
	"github.com/stretchr/testify/assert"
)

func (c *ColumnsTestSuite) TestShouldSkipColumn() {
	type _testCase struct {
		name                  string
		colName               string
		softDelete            bool
		includeArtieUpdatedAt bool
		expectedResult        bool
	}

	testCases := []_testCase{
		{
			name:       "delete col marker + soft delete",
			colName:    constants.DeleteColumnMarker,
			softDelete: true,
		},
		{
			name:           "delete col marker",
			colName:        constants.DeleteColumnMarker,
			expectedResult: true,
		},
		{
			name:                  "updated col marker + include updated",
			colName:               constants.UpdateColumnMarker,
			includeArtieUpdatedAt: true,
		},
		{
			name:           "updated col marker",
			colName:        constants.UpdateColumnMarker,
			expectedResult: true,
		},
		{
			name:    "random col",
			colName: "firstName",
		},
		{
			name:                  "col with includeArtieUpdatedAt + softDelete",
			colName:               "email",
			includeArtieUpdatedAt: true,
			softDelete:            true,
		},
	}

	for _, testCase := range testCases {
		actualResult := shouldSkipColumn(testCase.colName, testCase.softDelete, testCase.includeArtieUpdatedAt)
		assert.Equal(c.T(), testCase.expectedResult, actualResult, testCase.name)
	}
}

func (c *ColumnsTestSuite) TestDiff_VariousNils() {
	type _testCase struct {
		name       string
		sourceCols *Columns
		targCols   *Columns

		expectedSrcKeyLength  int
		expectedTargKeyLength int
	}

	var sourceColsNotNil Columns
	var targColsNotNil Columns
	sourceColsNotNil.AddColumn(NewColumn("foo", typing.Invalid))
	targColsNotNil.AddColumn(NewColumn("foo", typing.Invalid))
	testCases := []_testCase{
		{
			name:       "both &Columns{}",
			sourceCols: &Columns{},
			targCols:   &Columns{},
		},
		{
			name:                  "only targ is &Columns{}",
			sourceCols:            &sourceColsNotNil,
			targCols:              &Columns{},
			expectedTargKeyLength: 1,
		},
		{
			name:                 "only source is &Columns{}",
			sourceCols:           &Columns{},
			targCols:             &targColsNotNil,
			expectedSrcKeyLength: 1,
		},
		{
			name:       "both nil",
			sourceCols: nil,
			targCols:   nil,
		},
		{
			name:                  "only targ is nil",
			sourceCols:            &sourceColsNotNil,
			targCols:              nil,
			expectedTargKeyLength: 1,
		},
		{
			name:                 "only source is nil",
			sourceCols:           nil,
			targCols:             &targColsNotNil,
			expectedSrcKeyLength: 1,
		},
	}

	for _, testCase := range testCases {
		actualSrcKeysMissing, actualTargKeysMissing := Diff(c.ctx, testCase.sourceCols, testCase.targCols, false, false)
		assert.Equal(c.T(), testCase.expectedSrcKeyLength, len(actualSrcKeysMissing), testCase.name)
		assert.Equal(c.T(), testCase.expectedTargKeyLength, len(actualTargKeysMissing), testCase.name)
	}
}

func (c *ColumnsTestSuite) TestDiffBasic() {
	var source Columns
	source.AddColumn(NewColumn("a", typing.Integer))

	srcKeyMissing, targKeyMissing := Diff(c.ctx, &source, &source, false, false)
	assert.Equal(c.T(), len(srcKeyMissing), 0)
	assert.Equal(c.T(), len(targKeyMissing), 0)
}

func (c *ColumnsTestSuite) TestDiffDelta1() {
	var sourceCols Columns
	var targCols Columns
	for colName, kindDetails := range map[string]typing.KindDetails{
		"a": typing.String,
		"b": typing.Boolean,
		"c": typing.Struct,
	} {
		sourceCols.AddColumn(NewColumn(colName, kindDetails))
	}

	for colName, kindDetails := range map[string]typing.KindDetails{
		"aa": typing.String,
		"b":  typing.Boolean,
		"cc": typing.String,
	} {
		targCols.AddColumn(NewColumn(colName, kindDetails))
	}

	srcKeyMissing, targKeyMissing := Diff(c.ctx, &sourceCols, &targCols, false, false)
	assert.Equal(c.T(), len(srcKeyMissing), 2, srcKeyMissing)   // Missing aa, cc
	assert.Equal(c.T(), len(targKeyMissing), 2, targKeyMissing) // Missing aa, cc
}

func (c *ColumnsTestSuite) TestDiffDelta2() {
	var sourceCols Columns
	var targetCols Columns

	for colName, kindDetails := range map[string]typing.KindDetails{
		"a":  typing.String,
		"aa": typing.String,
		"b":  typing.Boolean,
		"c":  typing.Struct,
		"d":  typing.String,
		"CC": typing.String,
		"cC": typing.String,
		"Cc": typing.String,
	} {
		sourceCols.AddColumn(NewColumn(colName, kindDetails))
	}

	for colName, kindDetails := range map[string]typing.KindDetails{
		"aa": typing.String,
		"b":  typing.Boolean,
		"cc": typing.String,
		"CC": typing.String,
		"dd": typing.String,
	} {
		targetCols.AddColumn(NewColumn(colName, kindDetails))
	}

	srcKeyMissing, targKeyMissing := Diff(c.ctx, &sourceCols, &targetCols, false, false)
	assert.Equal(c.T(), len(srcKeyMissing), 1, srcKeyMissing)   // Missing dd
	assert.Equal(c.T(), len(targKeyMissing), 3, targKeyMissing) // Missing a, c, d
}

func (c *ColumnsTestSuite) TestDiffDeterministic() {
	retMap := map[string]bool{}

	var sourceCols Columns
	var targCols Columns

	sourceCols.AddColumn(NewColumn("id", typing.Integer))
	sourceCols.AddColumn(NewColumn("name", typing.String))

	for i := 0; i < 500; i++ {
		keysMissing, targetKeysMissing := Diff(c.ctx, &sourceCols, &targCols, false, false)
		assert.Equal(c.T(), 0, len(keysMissing), keysMissing)

		var key string
		for _, targetKeyMissing := range targetKeysMissing {
			key += targetKeyMissing.Name(c.ctx, nil)
		}

		retMap[key] = false
	}

	assert.Equal(c.T(), 1, len(retMap), retMap)
}

func (c *ColumnsTestSuite) TestCopyColMap() {
	var cols Columns
	cols.AddColumn(NewColumn("hello", typing.String))
	cols.AddColumn(NewColumn("created_at", typing.NewKindDetailsFromTemplate(typing.ETime, ext.DateTimeKindType)))
	cols.AddColumn(NewColumn("updated_at", typing.NewKindDetailsFromTemplate(typing.ETime, ext.DateTimeKindType)))

	copiedCols := CloneColumns(&cols)
	assert.Equal(c.T(), *copiedCols, cols)

	//Delete a row from copiedCols
	copiedCols.columns = append(copiedCols.columns[1:])
	assert.NotEqual(c.T(), *copiedCols, cols)
}

func (c *ColumnsTestSuite) TestCloneColumns() {
	type _testCase struct {
		name         string
		cols         *Columns
		expectedCols *Columns
	}

	var cols Columns
	cols.AddColumn(NewColumn("foo", typing.String))
	cols.AddColumn(NewColumn("bar", typing.String))
	cols.AddColumn(NewColumn("xyz", typing.String))
	cols.AddColumn(NewColumn("abc", typing.String))

	var mixedCaseCols Columns
	mixedCaseCols.AddColumn(NewColumn("foo", typing.String))
	mixedCaseCols.AddColumn(NewColumn("bAr", typing.String))
	mixedCaseCols.AddColumn(NewColumn("XYZ", typing.String))
	mixedCaseCols.AddColumn(NewColumn("aBC", typing.String))

	testCases := []_testCase{
		{
			name:         "nil col",
			expectedCols: &Columns{},
		},
		{
			name:         "&Columns{}",
			cols:         &Columns{},
			expectedCols: &Columns{},
		},
		{
			name:         "copying columns",
			cols:         &cols,
			expectedCols: &cols,
		},
		{
			name:         "mixed case cols",
			cols:         &mixedCaseCols,
			expectedCols: &cols,
		},
	}

	for _, testCase := range testCases {
		actualCols := CloneColumns(testCase.cols)
		assert.Equal(c.T(), *testCase.expectedCols, *actualCols, testCase.name)
	}
}

package dml

import (
	"fmt"
	"strings"
	"time"

	"github.com/artie-labs/transfer/lib/sql"

	"github.com/artie-labs/transfer/lib/typing/columns"

	"github.com/artie-labs/transfer/lib/config/constants"
	"github.com/artie-labs/transfer/lib/typing"
	"github.com/stretchr/testify/assert"
)

func (m *MergeTestSuite) TestMergeStatementSoftDelete() {
	// No idempotent key
	fqTable := "database.schema.table"
	cols := []string{
		"id",
		"bar",
		"updated_at",
		constants.DeleteColumnMarker,
	}

	tableValues := []string{
		fmt.Sprintf("('%s', '%s', '%v', false)", "1", "456", time.Now().Round(0).Format(time.RFC3339)),
		fmt.Sprintf("('%s', '%s', '%v', true)", "2", "bb", time.Now().Round(0).Format(time.RFC3339)), // Delete row 2.
		fmt.Sprintf("('%s', '%s', '%v', false)", "3", "dd", time.Now().Round(0).Format(time.RFC3339)),
	}

	// select cc.foo, cc.bar from (values (12, 34), (44, 55)) as cc(foo, bar);
	subQuery := fmt.Sprintf("SELECT %s from (values %s) as %s(%s)",
		strings.Join(cols, ","), strings.Join(tableValues, ","), "_tbl", strings.Join(cols, ","))

	var _cols columns.Columns
	_cols.AddColumn(columns.NewColumn("id", typing.String))
	_cols.AddColumn(columns.NewColumn(constants.DeleteColumnMarker, typing.Boolean))

	for _, idempotentKey := range []string{"", "updated_at"} {
		mergeSQL, err := MergeStatement(m.ctx, &MergeArgument{
			FqTableName:    fqTable,
			SubQuery:       subQuery,
			IdempotentKey:  idempotentKey,
			PrimaryKeys:    []columns.Wrapper{columns.NewWrapper(m.ctx, columns.NewColumn("id", typing.Invalid), nil)},
			ColumnsToTypes: _cols,
			DestKind:       constants.Snowflake,
			SoftDelete:     true,
		})
		assert.NoError(m.T(), err)
		assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("MERGE INTO %s", fqTable)), mergeSQL)
		// Soft deletion flag being passed.
		assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("%s=cc.%s", constants.DeleteColumnMarker, constants.DeleteColumnMarker)), mergeSQL)

		assert.Equal(m.T(), len(idempotentKey) > 0, strings.Contains(mergeSQL, fmt.Sprintf("cc.%s >= c.%s", "updated_at", "updated_at")))
	}

}

func (m *MergeTestSuite) TestMergeStatement() {
	// No idempotent key
	fqTable := "database.schema.table"
	colToTypes := map[string]typing.KindDetails{
		"id":                         typing.String,
		"bar":                        typing.String,
		"updated_at":                 typing.String,
		"start":                      typing.String,
		constants.DeleteColumnMarker: typing.Boolean,
	}

	// This feels a bit round about, but this is because iterating over a map is not deterministic.
	cols := []string{"id", "bar", "updated_at", "start", constants.DeleteColumnMarker}
	var _cols columns.Columns
	for _, col := range cols {
		_cols.AddColumn(columns.NewColumn(col, colToTypes[col]))
	}

	tableValues := []string{
		fmt.Sprintf("('%s', '%s', '%v', '%v', false)", "1", "456", "foo", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%v', '%v', false)", "2", "bb", "bar", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%v', '%v', false)", "3", "dd", "world", time.Now().Round(0).UTC()),
	}

	// select cc.foo, cc.bar from (values (12, 34), (44, 55)) as cc(foo, bar);
	subQuery := fmt.Sprintf("SELECT %s from (values %s) as %s(%s)",
		strings.Join(cols, ","), strings.Join(tableValues, ","), "_tbl", strings.Join(cols, ","))
	mergeSQL, err := MergeStatement(m.ctx, &MergeArgument{
		FqTableName:    fqTable,
		SubQuery:       subQuery,
		IdempotentKey:  "",
		PrimaryKeys:    []columns.Wrapper{columns.NewWrapper(m.ctx, columns.NewColumn("id", typing.Invalid), nil)},
		ColumnsToTypes: _cols,
		DestKind:       constants.Snowflake,
		SoftDelete:     false,
	})
	assert.NoError(m.T(), err)
	assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("MERGE INTO %s", fqTable)), mergeSQL)
	assert.False(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("cc.%s >= c.%s", "updated_at", "updated_at")), fmt.Sprintf("Idempotency key: %s", mergeSQL))
	// Check primary keys clause
	assert.True(m.T(), strings.Contains(mergeSQL, "as cc on c.id = cc.id"), mergeSQL)

	// Check setting for update
	assert.True(m.T(), strings.Contains(mergeSQL, `SET id=cc.id,bar=cc.bar,updated_at=cc.updated_at,"start"=cc."start"`), mergeSQL)
	// Check for INSERT
	assert.True(m.T(), strings.Contains(mergeSQL, `id,bar,updated_at,"start"`), mergeSQL)
	assert.True(m.T(), strings.Contains(mergeSQL, `cc.id,cc.bar,cc.updated_at,cc."start"`), mergeSQL)
}

func (m *MergeTestSuite) TestMergeStatementIdempotentKey() {
	fqTable := "database.schema.table"
	cols := []string{
		"id",
		"bar",
		"updated_at",
		constants.DeleteColumnMarker,
	}

	tableValues := []string{
		fmt.Sprintf("('%s', '%s', '%v', false)", "1", "456", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%v', false)", "2", "bb", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%v', false)", "3", "dd", time.Now().Round(0).UTC()),
	}

	// select cc.foo, cc.bar from (values (12, 34), (44, 55)) as cc(foo, bar);
	subQuery := fmt.Sprintf("SELECT %s from (values %s) as %s(%s)",
		strings.Join(cols, ","), strings.Join(tableValues, ","), "_tbl", strings.Join(cols, ","))

	var _cols columns.Columns
	_cols.AddColumn(columns.NewColumn("id", typing.String))
	_cols.AddColumn(columns.NewColumn(constants.DeleteColumnMarker, typing.Boolean))

	mergeSQL, err := MergeStatement(m.ctx, &MergeArgument{
		FqTableName:    fqTable,
		SubQuery:       subQuery,
		IdempotentKey:  "updated_at",
		PrimaryKeys:    []columns.Wrapper{columns.NewWrapper(m.ctx, columns.NewColumn("id", typing.Invalid), nil)},
		ColumnsToTypes: _cols,
		DestKind:       constants.Snowflake,
		SoftDelete:     false,
	})
	assert.NoError(m.T(), err)
	assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("MERGE INTO %s", fqTable)), mergeSQL)
	assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("cc.%s >= c.%s", "updated_at", "updated_at")), fmt.Sprintf("Idempotency key: %s", mergeSQL))
}

func (m *MergeTestSuite) TestMergeStatementCompositeKey() {
	fqTable := "database.schema.table"
	cols := []string{
		"id",
		"another_id",
		"bar",
		"updated_at",
		constants.DeleteColumnMarker,
	}

	tableValues := []string{
		fmt.Sprintf("('%s', '%s', '%s', '%v', false)", "1", "3", "456", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%s', '%v', false)", "2", "2", "bb", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%s', '%v', false)", "3", "1", "dd", time.Now().Round(0).UTC()),
	}

	// select cc.foo, cc.bar from (values (12, 34), (44, 55)) as cc(foo, bar);
	subQuery := fmt.Sprintf("SELECT %s from (values %s) as %s(%s)",
		strings.Join(cols, ","), strings.Join(tableValues, ","), "_tbl", strings.Join(cols, ","))

	var _cols columns.Columns
	_cols.AddColumn(columns.NewColumn("id", typing.String))
	_cols.AddColumn(columns.NewColumn("another_id", typing.String))
	_cols.AddColumn(columns.NewColumn(constants.DeleteColumnMarker, typing.Boolean))

	mergeSQL, err := MergeStatement(m.ctx, &MergeArgument{
		FqTableName:   fqTable,
		SubQuery:      subQuery,
		IdempotentKey: "updated_at",
		PrimaryKeys: []columns.Wrapper{columns.NewWrapper(m.ctx, columns.NewColumn("id", typing.Invalid), nil),
			columns.NewWrapper(m.ctx, columns.NewColumn("another_id", typing.Invalid), nil)},
		ColumnsToTypes: _cols,
		DestKind:       constants.Snowflake,
		SoftDelete:     false,
	})
	assert.NoError(m.T(), err)
	assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("MERGE INTO %s", fqTable)), mergeSQL)
	assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("cc.%s >= c.%s", "updated_at", "updated_at")), fmt.Sprintf("Idempotency key: %s", mergeSQL))
	assert.True(m.T(), strings.Contains(mergeSQL, "cc on c.id = cc.id and c.another_id = cc.another_id"))
}

func (m *MergeTestSuite) TestMergeStatementEscapePrimaryKeys() {
	// No idempotent key
	fqTable := "database.schema.table"
	colToTypes := map[string]typing.KindDetails{
		"id":                         typing.String,
		"group":                      typing.String,
		"updated_at":                 typing.String,
		"start":                      typing.String,
		constants.DeleteColumnMarker: typing.Boolean,
	}

	// This feels a bit round about, but this is because iterating over a map is not deterministic.
	cols := []string{"id", "group", "updated_at", "start", constants.DeleteColumnMarker}
	var _cols columns.Columns
	for _, col := range cols {
		_cols.AddColumn(columns.NewColumn(col, colToTypes[col]))
	}

	tableValues := []string{
		fmt.Sprintf("('%s', '%s', '%v', '%v', false)", "1", "456", "foo", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%v', '%v', false)", "2", "bb", "bar", time.Now().Round(0).UTC()),
		fmt.Sprintf("('%s', '%s', '%v', '%v', false)", "3", "dd", "world", time.Now().Round(0).UTC()),
	}

	// select cc.foo, cc.bar from (values (12, 34), (44, 55)) as cc(foo, bar);
	subQuery := fmt.Sprintf("SELECT %s from (values %s) as %s(%s)",
		strings.Join(cols, ","), strings.Join(tableValues, ","), "_tbl", strings.Join(cols, ","))
	mergeSQL, err := MergeStatement(m.ctx, &MergeArgument{
		FqTableName:   fqTable,
		SubQuery:      subQuery,
		IdempotentKey: "",
		PrimaryKeys: []columns.Wrapper{
			columns.NewWrapper(m.ctx, columns.NewColumn("id", typing.Invalid), &sql.NameArgs{
				Escape:   true,
				DestKind: constants.Snowflake,
			}),
			columns.NewWrapper(m.ctx, columns.NewColumn("group", typing.Invalid), &sql.NameArgs{
				Escape:   true,
				DestKind: constants.Snowflake,
			}),
		},
		ColumnsToTypes: _cols,
		DestKind:       constants.Snowflake,
		SoftDelete:     false,
	})
	assert.NoError(m.T(), err)
	assert.True(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("MERGE INTO %s", fqTable)), mergeSQL)
	assert.False(m.T(), strings.Contains(mergeSQL, fmt.Sprintf("cc.%s >= c.%s", "updated_at", "updated_at")), fmt.Sprintf("Idempotency key: %s", mergeSQL))
	// Check primary keys clause
	assert.True(m.T(), strings.Contains(mergeSQL, `as cc on c.id = cc.id and c."group" = cc."group"`), mergeSQL)

	// Check setting for update
	assert.True(m.T(), strings.Contains(mergeSQL, `SET id=cc.id,"group"=cc."group",updated_at=cc.updated_at,"start"=cc."start"`), mergeSQL)
	// Check for INSERT
	assert.True(m.T(), strings.Contains(mergeSQL, `id,"group",updated_at,"start"`), mergeSQL)
	assert.True(m.T(), strings.Contains(mergeSQL, `cc.id,cc."group",cc.updated_at,cc."start"`), mergeSQL)
}

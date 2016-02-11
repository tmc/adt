package adt

type ColumnType uint8

const (
	ColumnTypeBool          ColumnType = 1
	ColumnTypeCharacter     ColumnType = 4
	ColumnTypeMemo          ColumnType = 5
	ColumnTypeBlob          ColumnType = 6
	ColumnTypeDouble        ColumnType = 10
	ColumnTypeInt           ColumnType = 11
	ColumnTypeShortInt      ColumnType = 12
	ColumnTypeCiCharacter   ColumnType = 20
	ColumnTypeAutoIncrement ColumnType = 15
	ColumnTypeDate          ColumnType = 3
	ColumnTypeTime          ColumnType = 13
	ColumnTypeTimestamp     ColumnType = 14
	ColumnTypeCurrency      ColumnType = 17
)

type MemoField struct {
	BlockOffset uint32
	Length      uint16
}

var sqlTypes = map[ColumnType]string{
	ColumnTypeBool:          "BOOLEAN",
	ColumnTypeCharacter:     "VARCHAR(255)",
	ColumnTypeMemo:          "TEXT",
	ColumnTypeBlob:          "BLOB",
	ColumnTypeDouble:        "DOUBLE",
	ColumnTypeInt:           "INTEGER",
	ColumnTypeShortInt:      "SMALLINT",
	ColumnTypeCiCharacter:   "VARCHAR(255)",
	ColumnTypeAutoIncrement: "INTEGER PRIMARY KEY", // AUTO_INCREMENT",
	ColumnTypeDate:          "DATE",
	ColumnTypeTime:          "TIME",
	ColumnTypeTimestamp:     "DATETIME",
	ColumnTypeCurrency:      "DECIMAL(10,2)",
}

func (ct ColumnType) SQLType() string {
	t, ok := sqlTypes[ct]
	if ok {
		return t
	}
	return "VARCHAR(100)"
}

package adt

type ColumnType uint8

const (
	ColumnTypeBool          ColumnType = 1
	ColumnTypeCharacter     ColumnType = 4
	ColumnTypeMemo          ColumnType = 5
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

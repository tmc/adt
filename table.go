package adt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	HeaderLength           = 400
	ColumnDescriptorLength = 200
	MagicHeader            = "Advantage Table"
)

var (
	ErrMagicHeaderNotFound = errors.New("adt: magic header missing")
)

type Table struct {
	RecordCount  uint32
	DataOffset   uint16
	RecordLength uint32
	Columns      []*Column
	data         io.ReadSeeker
	memoData     io.ReadSeeker
}

func TableFromPath(filePath string) (*Table, error) {
	adt, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(filePath)
	admPath := filePath[:len(filePath)-len(ext)] + ".ADM"
	adm, err := os.Open(admPath)
	if err != nil {
		return nil, err
	}
	return FromReaders(adt, adm)
}

func FromReaders(adtContent io.ReadSeeker, admContent io.ReadSeeker) (*Table, error) {
	header := make([]byte, HeaderLength)
	if _, err := io.ReadAtLeast(adtContent, header, HeaderLength); err != nil {
		return nil, err
	}
	if string(header[:len(MagicHeader)]) != MagicHeader {
		return nil, ErrMagicHeaderNotFound
	}
	table := &Table{
		Columns:  []*Column{},
		data:     adtContent,
		memoData: admContent,
	}
	if err := readLE(header[24:], &table.RecordCount); err != nil {
		return nil, err
	}
	if err := readLE(header[32:], &table.DataOffset); err != nil {
		return nil, err
	}
	if err := readLE(header[36:], &table.RecordLength); err != nil {
		return nil, err
	}
	for i := 0; i < table.columnCount(); i++ {
		c, err := ColumnFromReader(adtContent)
		if err != nil {
			return nil, err
		}
		table.Columns = append(table.Columns, c)
	}
	return table, nil
}

func (t *Table) columnCount() int {
	return int((t.DataOffset - HeaderLength) / 200)
}

func readLE(src []byte, dest interface{}) error {
	return binary.Read(bytes.NewReader(src), binary.LittleEndian, dest)
}

func (t *Table) Get(record int) (Record, error) {
	t.data.Seek(int64(int(t.DataOffset)+int(t.RecordLength)*record), 0)
	return t.readRecord()
}

func (t *Table) readRecord() (Record, error) {
	r := Record{}
	bytes := make([]byte, t.RecordLength)
	if _, err := io.ReadFull(t.data, bytes); err != nil {
		return nil, err
	}
	if string(bytes[:len(RecordMagicHeader)]) != RecordMagicHeader {
		return nil, ErrMagicHeaderNotFound
	}
	for _, column := range t.Columns {
		value, err := ReadValue(bytes, column)

		if asMemo, ok := value.(MemoField); ok {
			t.memoData.Seek(int64(asMemo.BlockOffset)*8, 0)
			data := make([]byte, asMemo.Length)
			if _, err := io.ReadFull(t.memoData, data); err != nil {
				return nil, err
			}
			value = string(data)
		}

		// dbg:
		//valueBytes := bytes[column.Offset : column.Offset+column.Length]
		//spew.Dump(column, valueBytes, value)
		if err != nil {
			return nil, err
		}
		r[column.Name] = value
	}
	return r, nil
}

func ReadValue(src []byte, column *Column) (interface{}, error) {
	valueBytes := src[column.Offset : column.Offset+column.Length]
	switch column.Type {
	case ColumnTypeCharacter:
		return string(valueBytes), nil
	case ColumnTypeShortInt:
		var value int16
		err := binary.Read(bytes.NewReader(valueBytes), binary.BigEndian, &value)
		return value, err
	case ColumnTypeInt:
		var value int32
		err := binary.Read(bytes.NewReader(valueBytes), binary.BigEndian, &value)
		return value, err
	case ColumnTypeMemo:
		var value MemoField
		err := binary.Read(bytes.NewReader(valueBytes), binary.LittleEndian, &value)
		return value, err
	default:
		value := make([]byte, column.Length)
		copy(value, src[column.Offset:column.Offset+column.Length])
		return value, nil
		return nil, fmt.Errorf("adt ReadValue: %s not implemented", column.Type)
	}
}

package adt

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
)

type Column struct {
	Name   string
	Type   ColumnType
	Length uint16
	Offset uint16
}

func ColumnFromReader(r io.Reader) (*Column, error) {
	header := make([]byte, ColumnDescriptorLength)
	if _, err := io.ReadAtLeast(r, header, ColumnDescriptorLength); err != nil {
		return nil, err
	}
	column := &Column{
		Name: strings.Trim(string(header[:128]), "\x00"),
	}
	var columnType uint16
	if err := binary.Read(bytes.NewBuffer(header[128:130]), binary.BigEndian, &columnType); err != nil {
		return nil, err
	}
	column.Type = ColumnType(columnType)
	if err := binary.Read(bytes.NewBuffer(header[130:132]), binary.BigEndian, &column.Offset); err != nil {
		return nil, err
	}
	if err := binary.Read(bytes.NewBuffer(header[134:136]), binary.BigEndian, &column.Length); err != nil {
		return nil, err
	}
	return column, nil
}

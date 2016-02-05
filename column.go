package adt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

var ErrReadingColumnDescriptor = errors.New("adt: error reading column descriptor")

type Column struct {
	Name          string
	Type          ColumnType
	Length        uint16
	Offset        uint16
	DecimalDigits uint16
}

func ColumnFromReader(r io.Reader) (*Column, error) {
	header := make([]byte, ColumnDescriptorLength)
	if _, err := io.ReadAtLeast(r, header, ColumnDescriptorLength); err != nil {
		if err == io.EOF {
			return nil, ErrReadingColumnDescriptor
		}
		return nil, err
	}
	column := &Column{
		Name: strings.Trim(string(header[:128]), "\x00"),
	}
	var columnType uint8
	if err := binary.Read(bytes.NewBuffer(header[129:130]), binary.LittleEndian, &columnType); err != nil {
		return nil, err
	}
	column.Type = ColumnType(columnType)
	if err := binary.Read(bytes.NewBuffer(header[131:133]), binary.LittleEndian, &column.Offset); err != nil {
		return nil, err
	}
	if err := binary.Read(bytes.NewBuffer(header[134:136]), binary.BigEndian, &column.Length); err != nil {
		return nil, err
	}
	if err := binary.Read(bytes.NewBuffer(header[138:140]), binary.BigEndian, &column.DecimalDigits); err != nil {
		return nil, err
	}
	return column, nil
}

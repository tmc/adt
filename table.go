package adt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	HeaderLength           = 400
	ColumnDescriptorLength = 200
	MagicHeader            = "Advantage Table"
)

var (
	ErrMagicHeaderNotFound = errors.New("adt: magic header missing")
	ErrMultiplePKs         = errors.New("adt: multiple primary keys")
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
	adm, _ := os.Open(admPath)
	// adm isn't required.
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

func (t *Table) GetPK() (*Column, error) {
	var result *Column
	for _, c := range t.Columns {
		if c.Type == ColumnTypeAutoIncrement {
			if result != nil {
				return nil, ErrMultiplePKs
			}
			result = c
		}
	}
	return result, nil
}

func (t *Table) Get(record int) (Record, error) {
	t.data.Seek(int64(int(t.DataOffset)+int(t.RecordLength)*record), 0)
	return t.readRecord()
}

func (t *Table) readRecord() (Record, error) {
	r := Record{}
	bytes := make([]byte, t.RecordLength)
	if r, err := io.ReadFull(t.data, bytes); err != nil {
		log.Warn("didn't read enough: ", r, err)
		return nil, err
	}
	if string(bytes[:len(RecordMagicHeader)]) != RecordMagicHeader {
		//return nil, ErrMagicHeaderNotFound
	}
	for _, column := range t.Columns {
		value, err := ReadValue(bytes, column)
		// dbg:
		//valueBytes := bytes[column.Offset : column.Offset+column.Length]

		if asMemo, ok := value.(MemoField); ok {
			t.memoData.Seek(int64(asMemo.BlockOffset)*8, 0)
			data := make([]byte, asMemo.Length)
			if _, err := io.ReadFull(t.memoData, data); err != nil {
				log.Warnln("didn't read enough for memo field", column.Name, err)
				return nil, nil
			}
			value = string(data)
		}

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
		return strings.Trim(string(valueBytes), " \u0000"), nil
	case ColumnTypeShortInt:
		var value int16
		err := binary.Read(bytes.NewReader(valueBytes), binary.LittleEndian, &value)
		// null encoded as minint
		if value == math.MinInt16 {
			return nil, nil
		}
		return value, err
	case ColumnTypeInt:
		var value int32
		err := binary.Read(bytes.NewReader(valueBytes), binary.LittleEndian, &value)
		if value == math.MinInt32 {
			return nil, nil
		}
		return value, err
	case ColumnTypeBlob:
		buf := src[column.Offset : column.Offset+column.Length]
		return buf, nil
	case ColumnTypeMemo:
		var value MemoField
		err := binary.Read(bytes.NewReader(valueBytes), binary.LittleEndian, &value)
		return value, err
	case ColumnTypeAutoIncrement:
		var value uint32
		err := binary.Read(bytes.NewReader(valueBytes), binary.LittleEndian, &value)
		return value, err
	case ColumnTypeBool:
		var value bool
		if src[column.Offset : column.Offset+column.Length][0] == 'T' {
			value = true
		}
		return value, nil
	case ColumnTypeTime:
		buf := src[column.Offset : column.Offset+column.Length]
		i := binary.LittleEndian.Uint32(buf)
		return time.Second * time.Duration(i), nil
	case ColumnTypeTimestamp:
		buf := src[column.Offset : column.Offset+column.Length]
		i := binary.LittleEndian.Uint32(buf[:4])
		j := binary.LittleEndian.Uint32(buf[4:])
		value := adtDatetimeToTime(int32(i), int32(j))
		if i == 0 {
			return nil, nil
		}
		return value, nil
	case ColumnTypeDate:
		buf := src[column.Offset : column.Offset+column.Length]
		i := binary.LittleEndian.Uint32(buf)
		value := adtDateToTime(int32(i))
		if i == 0 {
			return nil, nil
		}
		return value, nil
	case ColumnTypeCurrency:
		fallthrough
	case ColumnTypeDouble:
		buf := src[column.Offset : column.Offset+column.Length]
		var value float64
		err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &value)
		if value == -1.6e-322 {
			return nil, err
		}
		return value, err
	default:
		value := make([]byte, column.Length)
		copy(value, src[column.Offset:column.Offset+column.Length])
		return nil, fmt.Errorf("adt ReadValue: %s not implemented", column.Type)
		return value, nil
	}
}

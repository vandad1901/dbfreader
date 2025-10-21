package bytes

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vandad1901/dbfreader/types"
	"github.com/vandad1901/dbfreader/utils"
)

func ReadFromFile(fileName string) (*types.DBFFile, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", fileName, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(fmt.Errorf("closing file %s: %w", fileName, err))
		}
	}()

	dbfFile, err := ReadDBFile(f)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", fileName, err)
	}

	return dbfFile, nil
}

func ReadDBFile(f *os.File) (*types.DBFFile, error) {
	var (
		dbfFile = new(types.DBFFile)
		err     error
	)

	dbfFile.Header, err = ReadHeaderBytes(f)
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	dbfFile.ArrayDescriptors, err = ReadArrayDescriptorsBytes(f)
	if err != nil {
		return nil, fmt.Errorf("reading array descriptors: %w", err)
	}

	if len(dbfFile.ArrayDescriptors) != int((dbfFile.Header.NumberOfBytesOfHeader-32)/32) {
		return nil, fmt.Errorf("number of array descriptors read %d does not match expected %d",
			len(dbfFile.ArrayDescriptors),
			(dbfFile.Header.NumberOfBytesOfHeader-32)/32)
	}

	dbfFile.Records, err = ReadRecordsBytes(f, dbfFile)
	if err != nil {
		return nil, fmt.Errorf("reading records: %w", err)
	}

	return dbfFile, nil
}

func ReadHeaderBytes(f *os.File) (*types.Header, error) {
	type HeaderBytes struct {
		DBase                 int8     // bytes  0 -  0
		LastUpdateDate        [3]byte  // bytes  1 -  3
		NumberOfRecords       int32    // bytes  4 -  7
		NumberOfBytesOfHeader int16    // bytes  8 -  9
		NumberOfBytesOfRecord int16    // bytes 10 - 11
		_                     int16    // bytes 12 - 13
		IncompleteTransaction bool     // bytes 14 - 14
		EncryptionFlag        bool     // bytes 15 - 15
		_                     [12]byte // bytes 16 - 27
		ProductionMDXFileFlag bool     // bytes 28 - 28
		LanguageDriverID      int8     // bytes 29 - 29
		_                     [2]byte  // bytes 30 - 31
	}

	header := HeaderBytes{}

	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("reading header bytes: %w", err)
	}

	return &types.Header{
		DBase:                 header.DBase,
		LastUpdateDate:        10000*int32(header.LastUpdateDate[0]) + 100*int32(header.LastUpdateDate[1]) + int32(header.LastUpdateDate[2]),
		NumberOfRecords:       header.NumberOfRecords,
		NumberOfBytesOfHeader: header.NumberOfBytesOfHeader,
		NumberOfBytesOfRecord: header.NumberOfBytesOfRecord,
		IncompleteTransaction: header.IncompleteTransaction,
		EncryptionFlag:        header.EncryptionFlag,
		ProductionMDXFileFlag: header.ProductionMDXFileFlag,
		LanguageDriverID:      header.LanguageDriverID,
	}, nil
}

func ReadArrayDescriptorsBytes(f *os.File) ([]types.ArrayDescriptor, error) {
	type ArrayDescriptorBytes struct {
		FileName               [11]byte // bytes  0 - 10
		FieldType              byte     // bytes 11 - 11
		_                      [4]byte  // bytes 12 - 15
		FieldLength            uint8    // bytes 16 - 16
		FieldDecimalCount      uint8    // bytes 17 - 17
		WorkAreaID             uint16   // bytes 18 - 19
		Example                byte     // bytes 20 - 20
		_                      [10]byte // bytes 21 - 30
		ProductionMDXFieldFlag bool     // bytes 31 - 31
	}

	var arrayDescriptors []types.ArrayDescriptor
	for {
		currentOffset, nextChar, err := utils.PeakNextByte(f)
		if err != nil {
			return nil, fmt.Errorf("error in peeking next byte: %w", err)
		}

		if nextChar == types.ArrayDescriptorsTerminator {
			break
		}
		if _, err := f.Seek(-1, io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("error in seeking file at offset %d:\n\t %w", currentOffset, err)
		}

		var ad ArrayDescriptorBytes

		if err := binary.Read(f, binary.LittleEndian, &ad); err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("error in reading array descriptor at offset %d: %w", currentOffset, err)
		}
		arrayDescriptors = append(arrayDescriptors, types.ArrayDescriptor{
			FieldName:              strings.Trim(string(ad.FileName[:]), "\u0000 "),
			FieldType:              ad.FieldType,
			FieldLength:            ad.FieldLength,
			FieldDecimalCount:      ad.FieldDecimalCount,
			WorkAreaID:             ad.WorkAreaID,
			Example:                ad.Example,
			ProductionMDXFieldFlag: ad.ProductionMDXFieldFlag,
		})

	}

	return arrayDescriptors, nil
}

func ReadRecordsBytes(f *os.File, dbfFile *types.DBFFile) ([]map[string]types.Primitive, error) {
	records := make([]map[string]types.Primitive, int(dbfFile.Header.NumberOfRecords))
	for i := 0; i < int(dbfFile.Header.NumberOfRecords); i++ {
		var deletionFlag byte

		if err := binary.Read(f, binary.LittleEndian, &deletionFlag); err != nil {
			return nil, fmt.Errorf("reading record %d deletion flag: %w", i, err)
		}

		record := make(map[string]types.Primitive, len(dbfFile.ArrayDescriptors))
		for _, ad := range dbfFile.ArrayDescriptors {
			b := make([]byte, ad.FieldLength)
			if err := binary.Read(f, binary.LittleEndian, &b); err != nil {
				return nil, fmt.Errorf("reading record b %d field %s: %w", i, ad.FieldName, err)
			}

			value, err := ReadField(types.FieldType(ad.FieldType), b)
			if err != nil {
				return nil, fmt.Errorf("reading parsing %d field %s: %w", i, ad.FieldName, err)
			}

			record[ad.FieldName] = value
		}

		records[i] = record
	}

	return records, nil
}

func ReadField(fieldType types.FieldType, b []byte) (types.Primitive, error) {
	var value types.Primitive

	switch fieldType {
	case types.CharacterType:
		var c types.Character
		if err := c.ReadFromDBF(b); err != nil {
			return nil, err
		}
		value = c
	case types.DecimalType:
		var d types.Decimal
		if err := d.ReadFromDBF(b); err != nil {
			return nil, err
		}
		value = d
	case types.FloatType:
		var fl types.Float
		if err := fl.ReadFromDBF(b); err != nil {
			return nil, err
		}
		value = fl
	case types.LogicalType:
		var l types.Logical
		if err := l.ReadFromDBF(b); err != nil {
			return nil, err
		}
		value = l
	case types.MemoType:
		var m types.Memo
		if err := m.ReadFromDBF(b); err != nil {
			return nil, err
		}
		value = m
	case types.NumericType:
		var n types.Numeric
		if err := n.ReadFromDBF(b); err != nil {
			return nil, err
		}
		value = n
	default:
		return nil, fmt.Errorf("unrecognized type: %c", fieldType)
	}
	return value, nil
}

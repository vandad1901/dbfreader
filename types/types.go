package types

type FieldType rune

const ArrayDescriptorsTerminator = 0x0D

type ArrayDescriptor struct {
	FieldName              string
	FieldType              byte
	FieldLength            uint8
	FieldDecimalCount      uint8
	WorkAreaID             uint16
	Example                byte
	ProductionMDXFieldFlag bool
}

type Header struct {
	DBase                 int8
	LastUpdateDate        int32
	NumberOfRecords       int32
	NumberOfBytesOfHeader int16
	NumberOfBytesOfRecord int16
	IncompleteTransaction bool
	EncryptionFlag        bool
	ProductionMDXFileFlag bool
	LanguageDriverID      int8
}

type DBFFile struct {
	Header           *Header
	ArrayDescriptors []ArrayDescriptor
	Records          []map[string]Primitive
}

func (d DBFFile) recordBytesCount() int64 {
	var sum int64 = 0
	for _, ad := range d.ArrayDescriptors {
		sum = sum + int64(ad.FieldLength)
	}

	return sum
}

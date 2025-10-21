package utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func PeakNextByte(f *os.File) (int64, byte, error) {
	currentOffset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, 0, fmt.Errorf("error in getting current offset: %w", err)
	}

	var nextChar byte
	if err := binary.Read(f, binary.LittleEndian, &nextChar); err != nil {
		return 0, 0, fmt.Errorf("error in reading next byte at offset %d: %w", currentOffset, err)
	}

	return currentOffset, nextChar, nil
}

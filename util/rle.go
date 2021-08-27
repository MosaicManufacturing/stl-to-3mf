package util

import (
	"encoding/binary"
	"io/ioutil"
)

type Run struct {
	Length uint32
	Value uint8
}

type RLE []Run

func LoadRLE(path string) (RLE, error) {
	rleBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rle := make(RLE, 0)
	for i := 0; i < len(rleBytes); i += 5 {
		runLength := binary.LittleEndian.Uint32(rleBytes[i:])
		value := rleBytes[i + 4]
		rle = append(rle, Run{
			Length:  runLength,
			Value: value,
		})
	}
	return rle, nil
}


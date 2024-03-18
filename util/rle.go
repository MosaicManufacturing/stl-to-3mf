package util

import (
	"encoding/binary"
	"io/ioutil"
)

type Run struct {
	Length uint32
	Value  uint8
}

type RLE struct {
	Runs []Run
}

func LoadRLE(path string, filamentIds map[byte]byte) (*RLE, error) {
	rleBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rle := new(RLE)
	for i := 0; i < len(rleBytes); i += 5 {
		runLength := binary.LittleEndian.Uint32(rleBytes[i:])
		value := rleBytes[i+4]
		// map project input to filamentId for Element
		if filamentIds != nil {
			value = filamentIds[value]
		}
		rle.Runs = append(rle.Runs, Run{
			Length: runLength,
			Value:  value,
		})
	}
	return rle, nil
}

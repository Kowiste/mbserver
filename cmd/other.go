package main

import (
	"encoding/binary"
	"io/ioutil"
	"math"
	"os"
)

//ReadFile Read a File
func ReadFile(FilePath string) ([]byte, error) {
	file, err := os.Open(FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func doubleToByte(input []float64, pointer int) []uint16 {
	buf := make([]uint16, 0)
	n := math.Float64bits(input[pointer])
	buf = append(buf, uint16(n>>48))
	buf = append(buf, uint16(n>>32))
	buf = append(buf, uint16(n>>16))
	buf = append(buf, uint16(n))
	n = math.Float64bits(input[pointer+1])
	buf = append(buf, uint16(n>>48))
	buf = append(buf, uint16(n>>32))
	buf = append(buf, uint16(n>>16))
	buf = append(buf, uint16(n))
	return buf
}
func int16ToByte(input uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, input)
	return b
}

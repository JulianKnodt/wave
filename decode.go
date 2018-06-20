package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const wave_format_pcm = 0x001

type Wave struct {
	audioFormat   uint16
	numChannels   uint16
	sampleRate    uint32
	bitsPerSample uint16
	data          []byte
}

func main() {
	// temporary while I'm figuring things out
	file, err := os.Open("./sample.wav")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(Decode(file))
  println("Done!")
}

func readId(file io.Reader, id []byte) error {
	dst := make([]byte, len(id))
	n, err := file.Read(dst)
	if err != nil {
		return err
	}
	if n != len(id) || !bytes.Equal(dst, id) {
		return fmt.Errorf("%q did not match %q", string(dst), string(id))
	}
	return nil
}

func Decode(file io.Reader) (result Wave) {
	err := readId(file, []byte("RIFF"))
	if err != nil {
		panic(err)
	}
	dst := make([]byte, 4)
	_, err = file.Read(dst)
	if err != nil {
		panic(err)
	}
	chunkSize := binary.LittleEndian.Uint32(dst) // it's 4 bytes so 32 int
	err = readId(file, []byte("WAVE"))
	if err != nil {
		fmt.Println(chunkSize)
		panic(err)
	}
	for {
		_, err = file.Read(dst)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		switch string(dst) {
		case "fmt ":
			audioFmt, numChan, sampleRate, _, _, bitsPerSample, _ := readFmtSection(file)
			result.audioFormat = audioFmt
			result.numChannels = numChan
			result.sampleRate = sampleRate
			result.bitsPerSample = bitsPerSample
		case "data":
			data := readDataSection(file)
			result.data = data
		default:
			readUnknownChunk(file)
		}
	}
	return
}

func readUnknownChunk(file io.Reader) {
	// https://github.com/python/cpython/blob/master/Lib/chunk.py
	dst := make([]byte, 4)
	_, err := file.Read(dst)
	if err != nil {
		panic(err)
	}
	size := binary.LittleEndian.Uint32(dst) - 8
	dst = make([]byte, size)
  // data lies within this subchunk
	_, err = file.Read(dst)
}

func readFmtSection(file io.Reader) (af uint16, nc uint16, sr uint32, br uint32, ba uint16, bps uint16, extra []byte) {
	dst := make([]byte, 4)
	_, err := file.Read(dst)
	if err != nil {
		panic(err)
	}
	restSize := binary.LittleEndian.Uint32(dst)
	dst = make([]byte, restSize)
	_, err = file.Read(dst)
	if err != nil {
		panic(err)
	}
	b := binary.LittleEndian
	af = b.Uint16(dst[:2])     // audio format
	nc = b.Uint16(dst[2:4])    // Num Channels
	sr = b.Uint32(dst[4:8])    // Sample Rate
	br = b.Uint32(dst[8:12])   // Byte rate
	ba = b.Uint16(dst[12:14])  // Block Align
	bps = b.Uint16(dst[14:16]) // Bits per sample
	if restSize <= 16 {
		return
	}
	extraParamSize := b.Uint16(dst[16:18])
	extra = make([]byte, extraParamSize)
	copy(extra, dst[18:18+extraParamSize])
	return
}

func readDataSection(file io.Reader) []byte {
	dst := make([]byte, 4)
	_, err := file.Read(dst)
	if err != nil {
		panic(err)
	}
	size := binary.LittleEndian.Uint32(dst)
	dst = make([]byte, size)
	_, err = file.Read(dst)
	if err != nil {
		panic(err)
	}
	return dst
}

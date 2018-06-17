package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	Decode(file)
	fmt.Println("Done!")
}

func readId(file *os.File, id []byte) error {
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

func Decode(file *os.File) {
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
	audioFmt, numChan, sampRate, byteRate, _, bitsPerSample, _ := readFmtSection(file)
	fmt.Println(audioFmt)
	fmt.Println(numChan)
	fmt.Println(sampRate)
	fmt.Println(byteRate)
	fmt.Println(bitsPerSample)
	_ = readDataSection(file)
}

func readFmtSection(file *os.File) (af uint16, nc uint16, sr uint32, br uint32, ba uint16, bps uint16, extra []byte) {
	err := readId(file, []byte("fmt "))
	if err != nil {
		panic(err)
	}
	dst := make([]byte, 4)
	_, err = file.Read(dst)
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

func readDataSection(file *os.File) []byte {
	err := readId(file, []byte("data"))
	if err != nil {
		panic(err)
	}
	dst := make([]byte, 4)
	_, err = file.Read(dst)
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

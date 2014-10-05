package buffer

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"testing"
)

func TestFile(t *testing.T) {
	buf := NewFile(1024)
	checkCap(t, buf, 1024)
	runPerfectSeries(t, buf)
}

func TestMem(t *testing.T) {
	buf := New(1024)
	checkCap(t, buf, 1024)
	runPerfectSeries(t, buf)
}

func TestFilePartition(t *testing.T) {
	buf := NewPartition(1024, NewFile)
	checkCap(t, buf, MaxCap())
	runPerfectSeries(t, buf)
}

func TestMulti(t *testing.T) {
	buf := NewMulti(New(5), New(5), NewFile(500), NewPartition(1024, New))
	checkCap(t, buf, MaxCap())
	runPerfectSeries(t, buf)
	isPerfectMatch(t, buf, 1024*1024)
}

func runPerfectSeries(t *testing.T, buf Buffer) {
	checkEmpty(t, buf)
	simple(t, buf)

	max := LimitAlloc(buf.Cap())
	isPerfectMatch(t, buf, 0)
	for i := int64(1); i < max; i *= 2 {
		isPerfectMatch(t, buf, i)
	}
	isPerfectMatch(t, buf, max)
}

func simple(t *testing.T, buf Buffer) {
	buf.Write([]byte("hello world"))
	data, _ := ioutil.ReadAll(buf)
	if !bytes.Equal([]byte("hello world"), data) {
		t.Error("Hello world failed.")
	}

	buf.Write([]byte("hello world"))
	data = make([]byte, 3)
	buf.Read(data)
	buf.Write([]byte(" yolo"))
	data, _ = ioutil.ReadAll(buf)
	if !bytes.Equal([]byte("lo world yolo"), data) {
		t.Error("Buffer crossing error :(")
	}
}

func backAndForth(t *testing.T, buf Buffer, size int64) {

	r := io.LimitReader(rand.Reader, size)
	tee := io.TeeReader(r, buf)

	wrote, _ := ioutil.ReadAll(tee)

	half := int64(512)
	halfRead := make([]byte, half)

	n, _ := buf.Read(halfRead)
	halfRead = halfRead[:n]
	halfWrote := wrote[:n]

	if !bytes.Equal(halfWrote, halfRead) {
		t.Error("Back and forth error")
	}

	buf.Write(wrote[n:])

	fullRead, _ := ioutil.ReadAll(buf)

	if !bytes.Equal(append(wrote[n:], wrote[n:]...), fullRead) {
		t.Error("Back and forth error 2")
	}

}

func buildOutputs(t *testing.T, buf Buffer, size int64) (wrote []byte, read []byte) {
	r := io.LimitReader(rand.Reader, size)
	tee := io.TeeReader(r, buf)

	wrote, _ = ioutil.ReadAll(tee)
	read, _ = ioutil.ReadAll(buf)

	return wrote, read
}

func isPerfectMatch(t *testing.T, buf Buffer, size int64) {
	wrote, read := buildOutputs(t, buf, size)
	if !bytes.Equal(wrote, read) {
		t.Error("Buffer should have matched")
	}

	backAndForth(t, buf, size)
}

func checkEmpty(t *testing.T, buf Buffer) {
	if buf.Len() != 0 {
		t.Error("Buffer should start empty!")
	}
}

func checkCap(t *testing.T, buf Buffer, correctCap int64) {
	if buf.Cap() != correctCap {
		t.Error("Buffer cap is incorrect", buf.Cap(), correctCap)
	}
}
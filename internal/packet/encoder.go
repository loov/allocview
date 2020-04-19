package packet

import "encoding/binary"

type Encoder struct {
	data []byte
}

func NewEncoder(size int) Encoder {
	enc := Encoder{make([]byte, 0, size)}
	enc.Reset()
	return enc
}

func (enc *Encoder) Reset() {
	enc.data = enc.data[:0]
	enc.Int32(0)
}

func (enc *Encoder) BytesLen() int {
	return len(enc.data) - 4
}

func (enc *Encoder) LengthAndBytes() []byte {
	binary.LittleEndian.PutUint32(enc.data, uint32(len(enc.data)-4))
	return enc.data
}

func (enc *Encoder) Byte(v byte) {
	enc.data = append(enc.data, v)
}

func (enc *Encoder) Uint32(v uint32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], v)
	enc.data = append(enc.data, data[:]...)
}

func (enc *Encoder) Uint64(v uint64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], v)
	enc.data = append(enc.data, data[:]...)
}

func (enc *Encoder) String(v string) {
	enc.Uint32(uint32(len(v)))
	enc.data = append(enc.data, []byte(v)...)
}

func (enc *Encoder) Int32(v int32) {
	enc.Uint32(uint32(v))
}

func (enc *Encoder) Int64(v int64) {
	enc.Uint64(uint64(v))
}

func (enc *Encoder) Uintptr(v uintptr) {
	enc.Uint64(uint64(v))
}

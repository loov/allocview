package packet

import "encoding/binary"

type Decoder struct {
	off  int
	data []byte
}

func (dec *Decoder) Reset(data []byte) {
	dec.off = 0
	dec.data = data
}

func (dec *Decoder) Byte() byte {
	dec.off++
	return dec.data[dec.off-1]
}

func (dec *Decoder) Uint32() uint32 {
	v := binary.LittleEndian.Uint32(dec.data[dec.off:])
	dec.off += 4
	return v
}

func (dec *Decoder) Uint64() uint64 {
	v := binary.LittleEndian.Uint64(dec.data[dec.off:])
	dec.off += 8
	return v
}

func (dec *Decoder) Int32() int32 {
	return int32(dec.Uint32())
}

func (dec *Decoder) Int64() int64 {
	return int64(dec.Uint64())
}

func (dec *Decoder) Uintptr() uintptr {
	return uintptr(dec.Uint64())
}

func (dec *Decoder) String() string {
	n := int(dec.Uint32())
	b := dec.data[dec.off : dec.off+n]
	dec.off += n
	return string(b)
}

package crawler

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *DbContent) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.ID, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "content":
			z.Content, err = dc.ReadBytes(z.Content)
			if err != nil {
				return
			}
		case "hash":
			err = dc.ReadExactBytes(z.Hash[:])
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *DbContent) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "id"
	err = en.Append(0x83, 0xa2, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.ID)
	if err != nil {
		return
	}
	// write "content"
	err = en.Append(0xa7, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBytes(z.Content)
	if err != nil {
		return
	}
	// write "hash"
	err = en.Append(0xa4, 0x68, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteBytes(z.Hash[:])
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *DbContent) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "id"
	o = append(o, 0x83, 0xa2, 0x69, 0x64)
	o = msgp.AppendUint64(o, z.ID)
	// string "content"
	o = append(o, 0xa7, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
	o = msgp.AppendBytes(o, z.Content)
	// string "hash"
	o = append(o, 0xa4, 0x68, 0x61, 0x73, 0x68)
	o = msgp.AppendBytes(o, z.Hash[:])
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DbContent) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.ID, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "content":
			z.Content, bts, err = msgp.ReadBytesBytes(bts, z.Content)
			if err != nil {
				return
			}
		case "hash":
			bts, err = msgp.ReadExactBytes(bts, z.Hash[:])
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *DbContent) Msgsize() (s int) {
	s = 1 + 3 + msgp.Uint64Size + 8 + msgp.BytesPrefixSize + len(z.Content) + 5 + msgp.ArrayHeaderSize + (16 * (msgp.ByteSize))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *DbMeta) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "lastId":
			z.LastID, err = dc.ReadUint64()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z DbMeta) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "lastId"
	err = en.Append(0x81, 0xa6, 0x6c, 0x61, 0x73, 0x74, 0x49, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.LastID)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z DbMeta) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "lastId"
	o = append(o, 0x81, 0xa6, 0x6c, 0x61, 0x73, 0x74, 0x49, 0x64)
	o = msgp.AppendUint64(o, z.LastID)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DbMeta) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "lastId":
			z.LastID, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z DbMeta) Msgsize() (s int) {
	s = 1 + 7 + msgp.Uint64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *DbURL) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.ID, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "count":
			z.Count, err = dc.ReadUint32()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z DbURL) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "id"
	err = en.Append(0x82, 0xa2, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.ID)
	if err != nil {
		return
	}
	// write "count"
	err = en.Append(0xa5, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.Count)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z DbURL) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "id"
	o = append(o, 0x82, 0xa2, 0x69, 0x64)
	o = msgp.AppendUint64(o, z.ID)
	// string "count"
	o = append(o, 0xa5, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	o = msgp.AppendUint32(o, z.Count)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DbURL) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.ID, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "count":
			z.Count, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z DbURL) Msgsize() (s int) {
	s = 1 + 3 + msgp.Uint64Size + 6 + msgp.Uint32Size
	return
}

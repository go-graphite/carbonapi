package responses

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *GraphiteFetchResponse) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "start":
			z.Start, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "end":
			z.End, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "step":
			z.Step, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "name":
			z.Name, err = dc.ReadString()
			if err != nil {
				return
			}
		case "pathExpression":
			z.PathExpression, err = dc.ReadString()
			if err != nil {
				return
			}
		case "values":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Values) >= int(zb0002) {
				z.Values = (z.Values)[:zb0002]
			} else {
				z.Values = make([]float64, zb0002)
			}
			for za0001 := range z.Values {
				z.Values[za0001], err = dc.ReadFloat64()
				if err != nil {
					return
				}
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
func (z *GraphiteFetchResponse) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "start"
	err = en.Append(0x86, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Start)
	if err != nil {
		return
	}
	// write "end"
	err = en.Append(0xa3, 0x65, 0x6e, 0x64)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.End)
	if err != nil {
		return
	}
	// write "step"
	err = en.Append(0xa4, 0x73, 0x74, 0x65, 0x70)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Step)
	if err != nil {
		return
	}
	// write "name"
	err = en.Append(0xa4, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Name)
	if err != nil {
		return
	}
	// write "pathExpression"
	err = en.Append(0xae, 0x70, 0x61, 0x74, 0x68, 0x45, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteString(z.PathExpression)
	if err != nil {
		return
	}
	// write "values"
	err = en.Append(0xa6, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Values)))
	if err != nil {
		return
	}
	for za0001 := range z.Values {
		err = en.WriteFloat64(z.Values[za0001])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *GraphiteFetchResponse) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "start"
	o = append(o, 0x86, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
	o = msgp.AppendUint32(o, z.Start)
	// string "end"
	o = append(o, 0xa3, 0x65, 0x6e, 0x64)
	o = msgp.AppendUint32(o, z.End)
	// string "step"
	o = append(o, 0xa4, 0x73, 0x74, 0x65, 0x70)
	o = msgp.AppendUint32(o, z.Step)
	// string "name"
	o = append(o, 0xa4, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Name)
	// string "pathExpression"
	o = append(o, 0xae, 0x70, 0x61, 0x74, 0x68, 0x45, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, z.PathExpression)
	// string "values"
	o = append(o, 0xa6, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Values)))
	for za0001 := range z.Values {
		o = msgp.AppendFloat64(o, z.Values[za0001])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *GraphiteFetchResponse) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "start":
			z.Start, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "end":
			z.End, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "step":
			z.Step, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "name":
			z.Name, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "pathExpression":
			z.PathExpression, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "values":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Values) >= int(zb0002) {
				z.Values = (z.Values)[:zb0002]
			} else {
				z.Values = make([]float64, zb0002)
			}
			for za0001 := range z.Values {
				z.Values[za0001], bts, err = msgp.ReadFloat64Bytes(bts)
				if err != nil {
					return
				}
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *GraphiteFetchResponse) Msgsize() (s int) {
	s = 1 + 6 + msgp.Uint32Size + 4 + msgp.Uint32Size + 5 + msgp.Uint32Size + 5 + msgp.StringPrefixSize + len(z.Name) + 15 + msgp.StringPrefixSize + len(z.PathExpression) + 7 + msgp.ArrayHeaderSize + (len(z.Values) * (msgp.Float64Size))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *GraphiteGlobResponse) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "isLeaf":
			z.IsLeaf, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "path":
			z.Path, err = dc.ReadString()
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
func (z GraphiteGlobResponse) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "isLeaf"
	err = en.Append(0x82, 0xa6, 0x69, 0x73, 0x4c, 0x65, 0x61, 0x66)
	if err != nil {
		return
	}
	err = en.WriteBool(z.IsLeaf)
	if err != nil {
		return
	}
	// write "path"
	err = en.Append(0xa4, 0x70, 0x61, 0x74, 0x68)
	if err != nil {
		return
	}
	err = en.WriteString(z.Path)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z GraphiteGlobResponse) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "isLeaf"
	o = append(o, 0x82, 0xa6, 0x69, 0x73, 0x4c, 0x65, 0x61, 0x66)
	o = msgp.AppendBool(o, z.IsLeaf)
	// string "path"
	o = append(o, 0xa4, 0x70, 0x61, 0x74, 0x68)
	o = msgp.AppendString(o, z.Path)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *GraphiteGlobResponse) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "isLeaf":
			z.IsLeaf, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "path":
			z.Path, bts, err = msgp.ReadStringBytes(bts)
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z GraphiteGlobResponse) Msgsize() (s int) {
	s = 1 + 7 + msgp.BoolSize + 5 + msgp.StringPrefixSize + len(z.Path)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MultiGraphiteFetchResponse) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(MultiGraphiteFetchResponse, zb0002)
	}
	for zb0001 := range *z {
		err = (*z)[zb0001].DecodeMsg(dc)
		if err != nil {
			return
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z MultiGraphiteFetchResponse) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zb0003 := range z {
		err = z[zb0003].EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MultiGraphiteFetchResponse) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0003 := range z {
		o, err = z[zb0003].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MultiGraphiteFetchResponse) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(MultiGraphiteFetchResponse, zb0002)
	}
	for zb0001 := range *z {
		bts, err = (*z)[zb0001].UnmarshalMsg(bts)
		if err != nil {
			return
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z MultiGraphiteFetchResponse) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0003 := range z {
		s += z[zb0003].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MultiGraphiteGlobResponse) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(MultiGraphiteGlobResponse, zb0002)
	}
	for zb0001 := range *z {
		var field []byte
		_ = field
		var zb0003 uint32
		zb0003, err = dc.ReadMapHeader()
		if err != nil {
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, err = dc.ReadMapKeyPtr()
			if err != nil {
				return
			}
			switch msgp.UnsafeString(field) {
			case "isLeaf":
				(*z)[zb0001].IsLeaf, err = dc.ReadBool()
				if err != nil {
					return
				}
			case "path":
				(*z)[zb0001].Path, err = dc.ReadString()
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
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z MultiGraphiteGlobResponse) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zb0004 := range z {
		// map header, size 2
		// write "isLeaf"
		err = en.Append(0x82, 0xa6, 0x69, 0x73, 0x4c, 0x65, 0x61, 0x66)
		if err != nil {
			return
		}
		err = en.WriteBool(z[zb0004].IsLeaf)
		if err != nil {
			return
		}
		// write "path"
		err = en.Append(0xa4, 0x70, 0x61, 0x74, 0x68)
		if err != nil {
			return
		}
		err = en.WriteString(z[zb0004].Path)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MultiGraphiteGlobResponse) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0004 := range z {
		// map header, size 2
		// string "isLeaf"
		o = append(o, 0x82, 0xa6, 0x69, 0x73, 0x4c, 0x65, 0x61, 0x66)
		o = msgp.AppendBool(o, z[zb0004].IsLeaf)
		// string "path"
		o = append(o, 0xa4, 0x70, 0x61, 0x74, 0x68)
		o = msgp.AppendString(o, z[zb0004].Path)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MultiGraphiteGlobResponse) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(MultiGraphiteGlobResponse, zb0002)
	}
	for zb0001 := range *z {
		var field []byte
		_ = field
		var zb0003 uint32
		zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
		if err != nil {
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				return
			}
			switch msgp.UnsafeString(field) {
			case "isLeaf":
				(*z)[zb0001].IsLeaf, bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					return
				}
			case "path":
				(*z)[zb0001].Path, bts, err = msgp.ReadStringBytes(bts)
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
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z MultiGraphiteGlobResponse) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0004 := range z {
		s += 1 + 7 + msgp.BoolSize + 5 + msgp.StringPrefixSize + len(z[zb0004].Path)
	}
	return
}

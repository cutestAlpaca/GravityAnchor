// Package protobuf provides lightweight protobuf varint encoding/decoding
// without depending on google/protobuf. It supports the minimal subset of
// protobuf wire format needed to manipulate conversation sidebar index entries.
package protobuf

import (
	"errors"
	"fmt"
)

// Wire type constants as defined by the protobuf wire format specification.
const (
	WireVarint          = 0
	Wire64Bit           = 1
	WireLengthDelimited = 2
	Wire32Bit           = 5
)

// ErrInvalidVarint is returned when a varint encoding is malformed.
var ErrInvalidVarint = errors.New("protobuf: invalid varint encoding")

// EncodeVarint encodes a uint64 value into protobuf varint format.
// Varints use one or more bytes; each byte's MSB indicates whether more
// bytes follow.
func EncodeVarint(value uint64) []byte {
	if value == 0 {
		return []byte{0}
	}
	var buf [10]byte // max varint size for uint64
	n := 0
	for value > 0 {
		b := byte(value & 0x7F)
		value >>= 7
		if value > 0 {
			b |= 0x80
		}
		buf[n] = b
		n++
	}
	result := make([]byte, n)
	copy(result, buf[:n])
	return result
}

// DecodeVarint decodes a varint starting at position pos in data.
// It returns the decoded value, the new position after the varint, and any error.
func DecodeVarint(data []byte, pos int) (uint64, int, error) {
	if pos >= len(data) {
		return 0, pos, fmt.Errorf("protobuf: position %d out of bounds (len=%d)", pos, len(data))
	}

	var value uint64
	var shift uint
	startPos := pos

	for {
		if pos >= len(data) {
			return 0, startPos, ErrInvalidVarint
		}
		if shift >= 64 {
			return 0, startPos, fmt.Errorf("protobuf: varint too long at position %d", startPos)
		}

		b := data[pos]
		value |= uint64(b&0x7F) << shift
		pos++
		shift += 7

		if b&0x80 == 0 {
			break
		}
	}

	return value, pos, nil
}

// SkipField advances the position past a field's value based on its wire type.
// This is used when parsing protobuf data to skip over fields we don't care about.
func SkipField(data []byte, pos int, wireType int) (int, error) {
	switch wireType {
	case WireVarint:
		// Skip over the varint bytes
		for {
			if pos >= len(data) {
				return pos, fmt.Errorf("protobuf: unexpected end of data while skipping varint at pos %d", pos)
			}
			b := data[pos]
			pos++
			if b&0x80 == 0 {
				break
			}
		}
		return pos, nil

	case Wire64Bit:
		if pos+8 > len(data) {
			return pos, fmt.Errorf("protobuf: unexpected end of data while skipping 64-bit field at pos %d", pos)
		}
		return pos + 8, nil

	case WireLengthDelimited:
		length, newPos, err := DecodeVarint(data, pos)
		if err != nil {
			return pos, fmt.Errorf("protobuf: error reading length-delimited field length: %w", err)
		}
		endPos := newPos + int(length)
		if endPos > len(data) {
			return pos, fmt.Errorf("protobuf: length-delimited field extends past end of data (need %d, have %d)", endPos, len(data))
		}
		return endPos, nil

	case Wire32Bit:
		if pos+4 > len(data) {
			return pos, fmt.Errorf("protobuf: unexpected end of data while skipping 32-bit field at pos %d", pos)
		}
		return pos + 4, nil

	default:
		return pos, fmt.Errorf("protobuf: unknown wire type %d at pos %d", wireType, pos)
	}
}

// StripField removes all instances of a specific field number from serialized
// protobuf data. This is used to remove fields before re-encoding them with
// updated values (e.g., stripping old timestamp fields before writing new ones).
func StripField(data []byte, targetFieldNumber int) []byte {
	if len(data) == 0 {
		return data
	}

	result := make([]byte, 0, len(data))
	pos := 0

	for pos < len(data) {
		// Remember where this field tag starts
		fieldStart := pos

		// Read the field tag (varint)
		tag, newPos, err := DecodeVarint(data, pos)
		if err != nil {
			// If we can't parse, keep the remaining data as-is
			result = append(result, data[fieldStart:]...)
			break
		}

		fieldNumber := int(tag >> 3)
		wireType := int(tag & 0x07)

		// Skip over the field value
		endPos, err := SkipField(data, newPos, wireType)
		if err != nil {
			// If we can't skip, keep remaining data as-is
			result = append(result, data[fieldStart:]...)
			break
		}

		// Only keep the field if it's not the target
		if fieldNumber != targetFieldNumber {
			result = append(result, data[fieldStart:endPos]...)
		}

		pos = endPos
	}

	return result
}

// EncodeLengthDelimited encodes data as a protobuf length-delimited field
// with the given field number. The encoding is: field tag (varint) + length (varint) + data.
func EncodeLengthDelimited(fieldNumber int, data []byte) []byte {
	tag := EncodeVarint(uint64(fieldNumber<<3 | WireLengthDelimited))
	length := EncodeVarint(uint64(len(data)))

	result := make([]byte, 0, len(tag)+len(length)+len(data))
	result = append(result, tag...)
	result = append(result, length...)
	result = append(result, data...)
	return result
}

// EncodeStringField encodes a string as a protobuf length-delimited field.
// This is a convenience wrapper around EncodeLengthDelimited.
func EncodeStringField(fieldNumber int, value string) []byte {
	return EncodeLengthDelimited(fieldNumber, []byte(value))
}

// BuildTimestampFields builds protobuf timestamp sub-messages for fields 3, 7, and 10.
// Each is a sub-message (length-delimited) containing sub-field 1 (varint) = seconds since epoch.
// These fields represent created_at, updated_at, and last_activity timestamps
// in the sidebar index entry format.
func BuildTimestampFields(epochSeconds int64) []byte {
	// Build the inner sub-message: field 1, varint, epochSeconds
	innerTag := EncodeVarint(uint64(1<<3 | WireVarint))
	innerValue := EncodeVarint(uint64(epochSeconds))

	inner := make([]byte, 0, len(innerTag)+len(innerValue))
	inner = append(inner, innerTag...)
	inner = append(inner, innerValue...)

	// Wrap as fields 3, 7, and 10
	var result []byte
	for _, fieldNum := range []int{3, 7, 10} {
		result = append(result, EncodeLengthDelimited(fieldNum, inner)...)
	}
	return result
}

// HasTimestampFields checks if the given protobuf blob contains any of the
// timestamp fields (3, 7, or 10). This is used to determine whether an
// existing entry already has timestamps or if they need to be added.
func HasTimestampFields(innerBlob []byte) bool {
	if len(innerBlob) == 0 {
		return false
	}

	pos := 0
	for pos < len(innerBlob) {
		tag, newPos, err := DecodeVarint(innerBlob, pos)
		if err != nil {
			return false
		}

		fieldNumber := int(tag >> 3)
		wireType := int(tag & 0x07)

		if fieldNumber == 3 || fieldNumber == 7 || fieldNumber == 10 {
			return true
		}

		endPos, err := SkipField(innerBlob, newPos, wireType)
		if err != nil {
			return false
		}
		pos = endPos
	}
	return false
}

package json_field

import "time"

type (
	Int    = JsonField[int]
	Int8   = JsonField[int8]
	Int16  = JsonField[int16]
	Int32  = JsonField[int32]
	Int64  = JsonField[int64]
	Uint   = JsonField[uint]
	Uint8  = JsonField[uint8]
	Uint16 = JsonField[uint16]
	Uint32 = JsonField[uint32]
	Uint64 = JsonField[uint64]

	Float32 = JsonField[float32]
	Float64 = JsonField[float64]

	Complex64  = JsonField[complex64]
	Complex128 = JsonField[complex128]

	String = JsonField[string]
	Byte   = JsonField[byte]
	Rune   = JsonField[rune]

	Bool = JsonField[bool]

	Time = JsonField[time.Time]
)

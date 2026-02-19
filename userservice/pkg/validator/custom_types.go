package validator

import (
	jsonField "github.com/Interpuls/ifc2-service-farm/pkg/json_field"
	"github.com/go-playground/validator/v10"
)

func registerBasicCustomTypes(v *validator.Validate) {
	v.RegisterCustomTypeFunc(jsonField.Valuer,
		jsonField.String{},
		jsonField.Int{},
		jsonField.Int8{},
		jsonField.Int16{},
		jsonField.Int32{},
		jsonField.Int64{},
		jsonField.Uint{},
		jsonField.Uint8{},
		jsonField.Uint16{},
		jsonField.Uint32{},
		jsonField.Uint64{},
		jsonField.Float32{},
		jsonField.Float64{},
		jsonField.Bool{},
		jsonField.Time{},
	)
}

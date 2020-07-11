package core

import (
	"github.com/bom-maker/bomcsv"
	"github.com/bom-maker/mouser/model"
)

// UberPart is representing a Mouser Part with information from the input BOM
type UberPart struct {
	model.Part
	bomcsv.CSVPart
}

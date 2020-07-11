package bomcsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/bom-maker/mouser/model"
)

// CSVPart represents an electronic part parsed from a CSV format
type CSVPart struct {
	// Quantity is the number of this part in the BOM
	Quantity model.APIUint
	// Value is the value of the part (Farad, Ohms, etc.)
	Value string
	// Device is the device name used in the CAD library
	Device string
	// MouserRef is the reference used from Mouser catalog
	MouserRef string
	// Package is the package name of the part
	Package string
	// Parts' is the name of the part used in the schematic
	Parts string
	// Description is the associated description of the part
	Description string
}

// setPartField set the part's field corresponding to the given header's name and associated value
func (p *CSVPart) setPartField(header, value string) error {
	switch header {
	case "Qty":
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		p.Quantity = model.APIUint(i)
	case "Value":
		p.Value = value
	case "Device":
		p.Device = value
	case "MouserRef":
		p.MouserRef = value
	case "Package":
		p.Package = value
	case "Parts":
		p.Parts = value
	case "Description":
		p.Description = value
	default:
		return fmt.Errorf("Ignoring unknown header %s", header)
	}

	return nil
}

// ReadCSVPartsFrom reads CSV from r and returns a slice of CSVPart
func ReadCSVPartsFrom(r io.Reader, sep string) ([]CSVPart, error) {
	parts := make([]CSVPart, 0)

	reader := csv.NewReader(r)
	reader.Comma, _ = utf8.DecodeRune([]byte(sep))

	// Reading header
	headers, err := reader.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read record from CSV, err=%+v", err)
	}

	for {
		record, err := reader.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return parts, fmt.Errorf("Unable to read record from CSV, err=%+v", err)
		}

		// Reading all fields and create a csv part from it
		csvPart := CSVPart{}
		for k, v := range record {
			err := csvPart.setPartField(headers[k], v)
			if err != nil {
				// Silently ignoring errors, field does not exist
				continue
			}
		}

		// Only store parts a Mouser reference is present
		if csvPart.MouserRef != "" {
			parts = append(parts, csvPart)
		}
	}

	return parts, nil
}

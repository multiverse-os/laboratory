package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/bom-maker/core"
)

var (
	_ Writer = (*CSV)(nil)
)

// CSV outputs as CSV format
type CSV struct {
	Parts     []core.UberPart
	Separator string
}

func (o *CSV) Write(w io.Writer) error {
	writer := csv.NewWriter(w)
	writer.Comma, _ = utf8.DecodeRune([]byte(o.Separator))

	for _, v := range o.Parts {
		record := []string{
			fmt.Sprintf("%d", v.Quantity),
			v.Parts,
			v.Device,
			v.Value,
			v.MouserRef,
			v.GetUnitPrice(1).Price,
			v.DatasheetURL,
			v.ProductDetailURL,
		}

		err := writer.Write(record)
		if err != nil {
			return fmt.Errorf("Cannot write record, err=%+v", err)
		}
		writer.Flush()
	}

	return nil
}

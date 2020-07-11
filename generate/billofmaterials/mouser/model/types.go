package model

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// APIUint is an unsigned integer number represented in the Mouser API
type APIUint uint

// APIFloat is a float64 represented in the Mouser API
type APIFloat float64

// GetAPIUintFromString returns an APIUint from a given string
func GetAPIUintFromString(s string) APIUint {
	v, err := strconv.ParseUint(parseNumberFromString(s), 10, 16)
	if err != nil {
		return APIUint(0)
	}
	return APIUint(v)
}

// GetAPIFloatFromString returns an APIFloat from a given string
func GetAPIFloatFromString(s string) APIFloat {
	v, err := strconv.ParseFloat(parseNumberFromString(s), 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing float64, err=%v", err)
		return APIFloat(0)
	}
	return APIFloat(v)
}

func parseNumberFromString(s string) string {
	re := regexp.MustCompile(`[0-9,\.]+`)
	return strings.ReplaceAll(re.FindString(s), ",", ".")
}

// UnmarshalJSON unmarshals a APIUint object type
func (i *APIUint) UnmarshalJSON(data []byte) error {
	// Basically treat input data as string and convert it to go type
	// Parameters can be:
	//   - quoted
	//   - unquoted
	//   - empty quoted
	//   - null
	var dataStr string
	s, err := strconv.Unquote(string(data))
	if err != nil {
		// No quote present in data
		s = string(data)
	}
	s = strconv.Quote(s)

	//log.Printf("APIUint, s=%s", s)
	if err := json.Unmarshal([]byte(s), &dataStr); err != nil {
		return err
	}

	r, err := strconv.ParseUint(dataStr, 10, 32)
	if err != nil {
		*i = APIUint(0)
	} else {
		*i = APIUint(r)
	}

	return nil
}

// UnmarshalJSON unmarshals a APIFloat object type
func (f *APIFloat) UnmarshalJSON(data []byte) error {
	// Basically treat input data as string and convert it to go type
	// Parameters can be:
	//   - quoted
	//   - unquoted
	//   - empty quoted
	//   - null
	var dataStr string
	s, err := strconv.Unquote(string(data))
	if err != nil {
		// No quote present in data
		s = string(data)
	}
	s = strconv.Quote(s)

	//log.Printf("APIFloat, s=%s", s)
	if err := json.Unmarshal([]byte(s), &dataStr); err != nil {
		return err
	}

	r, err := strconv.ParseFloat(dataStr, 32)
	if err != nil {
		*f = APIFloat(0)
	} else {
		*f = APIFloat(r)
	}

	return nil
}

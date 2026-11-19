package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexString unmarshals a JSON string, number, or boolean into a Go string.
// Runscope's API returns assertion values as numbers for comparisons like
// equal_number (e.g. response_status == 200), while this provider's Terraform
// schema and older API fixtures treat value as a string.
type FlexString string

func (f FlexString) String() string {
	return string(f)
}

func (f FlexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

func (f *FlexString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*f = ""
		return nil
	}

	switch b[0] {
	case '"':
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = FlexString(s)
		return nil
	case 't', 'f':
		var v bool
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*f = FlexString(strconv.FormatBool(v))
		return nil
	default:
		// number (int/float) — preserve as the JSON number's textual form
		var n json.Number
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		if err := dec.Decode(&n); err != nil {
			return fmt.Errorf("FlexString: cannot unmarshal %s: %w", string(b), err)
		}
		*f = FlexString(n.String())
		return nil
	}
}

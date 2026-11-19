package schema

import (
	"encoding/json"
	"testing"
)

func TestFlexString_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"string", `"200"`, "200"},
		{"int", `200`, "200"},
		{"float", `200.5`, "200.5"},
		{"bool true", `true`, "true"},
		{"bool false", `false`, "false"},
		{"null", `null`, ""},
		{"uuid string", `"0d813889-46cf-47ce-9bac-66e58d3372ed"`, "0d813889-46cf-47ce-9bac-66e58d3372ed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got FlexString
			if err := json.Unmarshal([]byte(tc.in), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("got %q want %q", got.String(), tc.want)
			}
		})
	}
}

func TestStepAssertion_NumericValue(t *testing.T) {
	// Live Runscope API returns response_status assertion values as numbers.
	const payload = `
{
  "meta": {"status": "success"},
  "data": {
    "id": "d7363d46-2c07-42db-bd2e-54b37e0094cc",
    "step_type": "request",
    "method": "GET",
    "url": "http://example.com",
    "assertions": [
      {
        "comparison": "equal_number",
        "source": "response_status",
        "value": 200
      }
    ],
    "variables": [],
    "scripts": [],
    "before_scripts": [],
    "headers": {},
    "auth": {},
    "skipped": false,
    "note": ""
  },
  "error": null
}
`
	var resp StepGetRequestResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatalf("unmarshal step with numeric assertion value: %v", err)
	}
	if len(resp.Step.Assertions) != 1 {
		t.Fatalf("expected 1 assertion, got %d", len(resp.Step.Assertions))
	}
	if got := resp.Step.Assertions[0].Value.String(); got != "200" {
		t.Fatalf("assertion value: got %q want %q", got, "200")
	}
}

func TestStep_CreateResponse_StringValueStillWorks(t *testing.T) {
	var resp StepCreateRequestResponse
	if err := json.Unmarshal([]byte(runscopeStepCreateResponse), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Step) != 1 || len(resp.Step[0].Assertions) < 1 {
		t.Fatal("expected assertions in fixture")
	}
	if got := resp.Step[0].Assertions[0].Value.String(); got != "200" {
		t.Fatalf("got %q want 200", got)
	}
}

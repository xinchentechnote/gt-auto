package codec

import (
	"testing"

	szse_bin "github.com/xinchentechnote/fin-proto-szse-bin-go/messages"
)

// TestConvertMapToStructNestedSliceField verifies the nested message shapes
// produced by cross-sheet/cross-file @ references fill real protocol fields:
// a list of string-valued objects maps onto []*PartitionReport with the loose
// string-to-number conversion applied per element.
func TestConvertMapToStructNestedSliceField(t *testing.T) {
	msg := szse_bin.NewReportSynchronization()
	err := ConvertMapToStruct(map[string]interface{}{
		"PartitionReport": []interface{}{
			map[string]interface{}{"PartitionNo": "1", "ReportIndex": "100"},
			map[string]interface{}{"PartitionNo": "2", "ReportIndex": "200"},
		},
	}, msg)
	if err != nil {
		t.Fatalf("failed to fill nested slice: %v", err)
	}
	if len(msg.PartitionReport) != 2 {
		t.Fatalf("expected 2 partition reports, got %d", len(msg.PartitionReport))
	}
	if msg.PartitionReport[1].PartitionNo != 2 || msg.PartitionReport[1].ReportIndex != 200 {
		t.Fatalf("unexpected element values: %+v", msg.PartitionReport[1])
	}
}

// TestConvertMapToStructNestedStructField verifies a nested object fills a
// struct-typed field.
func TestConvertMapToStructNestedStructField(t *testing.T) {
	type inner struct {
		StopPx int64  `json:"StopPx"`
		Side   string `json:"Side"`
	}
	type parent struct {
		Nested inner `json:"Nested"`
	}
	var out parent
	err := ConvertMapToStruct(map[string]interface{}{
		"Nested": map[string]interface{}{"StopPx": "999", "Side": "1"},
	}, &out)
	if err != nil {
		t.Fatalf("failed to fill nested struct: %v", err)
	}
	if out.Nested.StopPx != 999 || out.Nested.Side != "1" {
		t.Fatalf("unexpected nested values: %+v", out.Nested)
	}
}

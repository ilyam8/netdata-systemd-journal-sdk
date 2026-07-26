package main

import (
	"bytes"
	"testing"
)

func TestNetflowV5WorkloadMetadata(t *testing.T) {
	metadata, ok := metadataForWorkload(netflowV5Workload)
	if !ok {
		t.Fatal("netflow workload metadata missing")
	}
	if metadata.applicationFieldsPerRow != 29 ||
		metadata.entryItemsPerRow != 30 ||
		metadata.applicationLogicalBytesPerRow != 450 ||
		metadata.totalLogicalDataBytesPerRow != 491 {
		t.Fatalf("unexpected metadata: %+v", metadata)
	}
}

func TestNetflowV5RowsMatchExactShape(t *testing.T) {
	rows := makeNetflowV5Rows(256)
	if len(rows) != 256 {
		t.Fatalf("rows = %d, want 256", len(rows))
	}

	distinctApplicationPayloads := make(map[string]struct{})
	sourceAddresses := make(map[string]struct{})
	destinationAddresses := make(map[string]struct{})
	inputInterfaces := make(map[string]struct{})
	for index, row := range rows {
		if len(row.Fields) != 30 || len(row.Payloads) != 30 {
			t.Fatalf("row %d items = %d/%d, want 30/30", index, len(row.Fields), len(row.Payloads))
		}
		if row.Fields[0].Name != netflowBootFieldName ||
			string(row.Fields[0].Value) != netflowBootFieldValue {
			t.Fatalf("row %d boot field = %s=%s", index, row.Fields[0].Name, row.Fields[0].Value)
		}

		totalBytes := 0
		applicationBytes := 0
		for item, payload := range row.Payloads {
			totalBytes += len(payload)
			if item > 0 {
				applicationBytes += len(payload)
				distinctApplicationPayloads[string(payload)] = struct{}{}
			}
		}
		if applicationBytes != 450 || totalBytes != 491 {
			t.Fatalf("row %d logical bytes = %d/%d, want 450/491", index, applicationBytes, totalBytes)
		}

		sourceAddresses[string(fieldValue(row, "SRC_ADDR"))] = struct{}{}
		destinationAddresses[string(fieldValue(row, "DST_ADDR"))] = struct{}{}
		inputInterfaces[string(fieldValue(row, "IN_IF"))] = struct{}{}
	}

	if len(distinctApplicationPayloads) != 385 {
		t.Fatalf("distinct application payloads = %d, want 385", len(distinctApplicationPayloads))
	}
	if len(sourceAddresses) != 100 || len(destinationAddresses) != 3 || len(inputInterfaces) != 256 {
		t.Fatalf(
			"cardinality src/dst/in = %d/%d/%d, want 100/3/256",
			len(sourceAddresses),
			len(destinationAddresses),
			len(inputInterfaces),
		)
	}
}

func TestNetflowV5RawAndStructuredRowsMatch(t *testing.T) {
	for index, row := range makeNetflowV5Rows(256) {
		if len(row.Fields) != len(row.Payloads) {
			t.Fatalf("row %d fields/payloads = %d/%d", index, len(row.Fields), len(row.Payloads))
		}
		for item, field := range row.Fields {
			expected := makePayload(field.Name, field.Value)
			if !bytes.Equal(row.Payloads[item], expected) {
				t.Fatalf("row %d item %d raw/structured payload mismatch", index, item)
			}
		}
	}
}

func fieldValue(row benchRow, name string) []byte {
	for _, field := range row.Fields {
		if field.Name == name {
			return field.Value
		}
	}
	return nil
}

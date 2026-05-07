package parser

import (
	"os"
	"testing"
)

func TestParser(t *testing.T) {
	input, err := os.Open("testdata/test.go")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParser("*stdin*", input)
	b, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if b == nil {
		t.Fatal("nil")
	}

	if b.TableName != "node" {
		t.Fatal("table name")
	}

	if b.Fields.Type != "*repr.NodeSummary" {
		t.Fatal("type name")
	}
	if len(b.Fields.Fields) != 8 {
		t.Fatal("Number of fields")
	}
	if b.Fields.Fields[0].Name != "Timestamp" {
		t.Fatal("Field 0 name")
	}
	if b.Fields.Fields[0].Type != "string" {
		t.Fatal("Field 0 type")
	}
	if len(b.Fields.Fields[0].Attrs) != 2 {
		t.Fatal("Field 0 attrs")
	}

	if b.Generate == "" || b.Generate != "NodeSomething" {
		t.Fatal("Generate")
	}

	if b.Summary == nil || b.Summary.Command != "NodeCommand" {
		t.Fatal("Summary")
	}

	if b.Help == nil || b.Help.Command != "NodeCommand" {
		t.Fatal("Help")
	}

	if len(b.Aliases) != 4 {
		t.Fatal("Aliases")
	}

	if b.Defaults == nil || len(b.Defaults) != 1 || b.Defaults[0] != "default" {
		t.Fatal("Defaults")
	}
}

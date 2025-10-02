package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// TypeInfo holds information about a struct type.
type TypeInfo struct {
	Package    string      `json:"package"`
	TypeName   string      `json:"typeName"`
	Fields     []FieldInfo `json:"fields"`
	EnumValues []EnumInfo  `json:"enumValues"`
	IsRoot     bool        `json:"isRoot"`

	DocString       string      `json:"docString"`
	ParsedDocString GoDocString `json:"parsedDocString"`
}

// FieldInfo holds information about a struct field.
type FieldInfo struct {
	FieldName      string   `json:"fieldName"`
	TypeName       string   `json:"typeName"`
	Package        string   `json:"package"`
	TypeDecorators []string `json:"typeDecorators"`

	DocString       string      `json:"docString"`
	ParsedDocString GoDocString `json:"parsedDocString"`
}

type EnumInfo struct {
	Name string `json:"name"`

	DocString       string      `json:"docString"`
	ParsedDocString GoDocString `json:"parsedDocString"`
}

func WriteJSON(allTypes map[string]TypeInfo, w io.Writer) error {
	// Sort the fields and enum values to ensure deterministic output.
	for _, typeInfo := range allTypes {
		sort.Slice(typeInfo.Fields, func(i, j int) bool {
			return typeInfo.Fields[i].FieldName < typeInfo.Fields[j].FieldName
		})
		sort.Slice(typeInfo.EnumValues, func(i, j int) bool {
			return typeInfo.EnumValues[i].Name < typeInfo.EnumValues[j].Name
		})
	}

	jsonData, err := json.MarshalIndent(allTypes, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	_, err = w.Write(jsonData)
	return err
}

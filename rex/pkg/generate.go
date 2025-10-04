package pkg

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
)

// GenerateDataJS writes the self-contained HTML to `w`.
func GenerateDataJS(types map[string]TypeInfo, w io.Writer, startType string) error {
	typeData, err := json.Marshal(types)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("const typeData = "))
	if err != nil {
		return err
	}
	_, err = w.Write(typeData)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(";\n"))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprintf("const startTypes = ['%s'];", startType)))
	if err != nil {
		return err
	}
	return nil
}

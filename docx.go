package dcdmaker

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
)

type DocxContent struct {
	DocumentXML string
	HasHeader   bool
	HasFooter   bool
}

func extractDocxContent(data []byte) (*DocxContent, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}

	var docXML strings.Builder
	content := &DocxContent{}

	for _, f := range r.File {
		switch {
		case f.Name == "word/document.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
			docXML.WriteString(buf.String())

		case strings.HasPrefix(f.Name, "word/header") && strings.HasSuffix(f.Name, ".xml"):
			content.HasHeader = true

		case strings.HasPrefix(f.Name, "word/footer") && strings.HasSuffix(f.Name, ".xml"):
			content.HasFooter = true
		}
	}

	if docXML.Len() == 0 {
		return nil, fmt.Errorf("word/document.xml not found in DOCX")
	}

	content.DocumentXML = docXML.String()
	return content, nil
}

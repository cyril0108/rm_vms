package utils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// PrintSimplifiedXML takes raw SOAP XML bytes and prints a clean,
// indented tree, stripping away distracting namespace URLs and prefixes.
func PrintSimplifiedXML(xmlData []byte) {
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	depth := 0

	fmt.Println("\n========== SIMPLIFIED XML DEBUG ==========")
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("XML Parsing Error: %v\n", err)
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			indent := strings.Repeat("  ", depth)
			
			// Extract meaningful attributes, explicitly ignore namespace definitions
			var attrs []string
			for _, attr := range t.Attr {
				// Drop the xmlns declarations entirely
				if attr.Name.Space != "xmlns" && attr.Name.Local != "xmlns" {
					attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attr.Name.Local, attr.Value))
				}
			}
			
			attrStr := ""
			if len(attrs) > 0 {
				attrStr = " " + strings.Join(attrs, " ")
			}

			// We use t.Name.Local to drop prefixes (turns <trt:Profiles> into <Profiles>)
			fmt.Printf("%s<%s%s>\n", indent, t.Name.Local, attrStr)
			depth++
			
		case xml.EndElement:
			depth--
			// Uncomment the next two lines if you want closing tags, 
			// but usually just seeing the indented structure is enough.
			// indent := strings.Repeat("  ", depth)
			// fmt.Printf("%s</%s>\n", indent, t.Name.Local)
			
		case xml.CharData:
			content := strings.TrimSpace(string(t))
			if content != "" {
				indent := strings.Repeat("  ", depth)
				// Print the actual values (like firmware version or IP)
				fmt.Printf("%s-> %s\n", indent, content)
			}
		}
	}
	fmt.Print("==========================================\n\n")
}
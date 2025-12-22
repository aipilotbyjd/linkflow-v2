package logic

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// XMLNode parses and generates XML
type XMLNode struct{}

func (n *XMLNode) Type() string {
	return "logic.xml"
}

func (n *XMLNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	operation := core.GetString(config, "operation", "parse")

	switch operation {
	case "parse":
		return n.parse(config, input)
	case "toJson":
		return n.toJSON(config, input)
	case "toXml":
		return n.toXML(config, input)
	case "xpath":
		return n.xpath(config, input)
	case "validate":
		return n.validate(config, input)
	default:
		return n.parse(config, input)
	}
}

func (n *XMLNode) parse(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	xmlData := getXMLInput(config, input)
	if xmlData == "" {
		return nil, fmt.Errorf("XML input is required")
	}

	result, err := parseXMLToMap(xmlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return map[string]interface{}{
		"data":   result,
		"parsed": true,
	}, nil
}

func (n *XMLNode) toJSON(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	xmlData := getXMLInput(config, input)
	if xmlData == "" {
		return nil, fmt.Errorf("XML input is required")
	}

	result, err := parseXMLToMap(xmlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	return map[string]interface{}{
		"json":   string(jsonBytes),
		"data":   result,
		"parsed": true,
	}, nil
}

func (n *XMLNode) toXML(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := config["data"]
	if data == nil {
		data = input["data"]
	}
	if data == nil {
		data = input
	}

	rootName := core.GetString(config, "rootName", "root")
	indent := core.GetBool(config, "indent", true)
	declaration := core.GetBool(config, "declaration", true)

	xmlStr, err := mapToXML(data, rootName, indent)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to XML: %w", err)
	}

	if declaration {
		xmlStr = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + xmlStr
	}

	return map[string]interface{}{
		"xml": xmlStr,
	}, nil
}

func (n *XMLNode) xpath(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	xmlData := getXMLInput(config, input)
	xpath := core.GetString(config, "xpath", "")

	if xmlData == "" {
		return nil, fmt.Errorf("XML input is required")
	}
	if xpath == "" {
		return nil, fmt.Errorf("XPath expression is required")
	}

	result, err := parseXMLToMap(xmlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Simple XPath navigation
	value := navigateXPath(result, xpath)

	return map[string]interface{}{
		"value": value,
		"found": value != nil,
	}, nil
}

func (n *XMLNode) validate(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	xmlData := getXMLInput(config, input)
	if xmlData == "" {
		return nil, fmt.Errorf("XML input is required")
	}

	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	var errors []string

	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, err.Error())
			break
		}
	}

	return map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	}, nil
}

// Helper functions

func getXMLInput(config map[string]interface{}, input map[string]interface{}) string {
	if xml := core.GetString(config, "xml", ""); xml != "" {
		return xml
	}
	if xml, ok := input["xml"].(string); ok {
		return xml
	}
	if xml, ok := input["body"].(string); ok {
		return xml
	}
	if xml, ok := input["data"].(string); ok {
		return xml
	}
	return ""
}

// XMLNode represents a parsed XML node
type xmlNode struct {
	Name       string
	Attrs      map[string]string
	Text       string
	Children   []*xmlNode
}

func parseXMLToMap(xmlData string) (map[string]interface{}, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	var stack []*xmlNode
	var root *xmlNode

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			node := &xmlNode{
				Name:  t.Name.Local,
				Attrs: make(map[string]string),
			}
			for _, attr := range t.Attr {
				node.Attrs[attr.Name.Local] = attr.Value
			}

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			} else {
				root = node
			}
			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(stack) > 0 {
				stack[len(stack)-1].Text = text
			}
		}
	}

	if root == nil {
		return make(map[string]interface{}), nil
	}

	return xmlNodeToMap(root), nil
}

func xmlNodeToMap(node *xmlNode) map[string]interface{} {
	result := make(map[string]interface{})

	// Add attributes with @ prefix
	for key, value := range node.Attrs {
		result["@"+key] = value
	}

	// Add text content
	if node.Text != "" && len(node.Children) == 0 {
		return map[string]interface{}{
			node.Name: node.Text,
		}
	}

	if node.Text != "" {
		result["#text"] = node.Text
	}

	// Group children by name
	childGroups := make(map[string][]*xmlNode)
	for _, child := range node.Children {
		childGroups[child.Name] = append(childGroups[child.Name], child)
	}

	// Convert children
	for name, children := range childGroups {
		if len(children) == 1 {
			childMap := xmlNodeToMap(children[0])
			if val, ok := childMap[name]; ok {
				result[name] = val
			} else {
				result[name] = childMap
			}
		} else {
			var arr []interface{}
			for _, child := range children {
				childMap := xmlNodeToMap(child)
				if val, ok := childMap[child.Name]; ok {
					arr = append(arr, val)
				} else {
					arr = append(arr, childMap)
				}
			}
			result[name] = arr
		}
	}

	return map[string]interface{}{node.Name: result}
}

func mapToXML(data interface{}, rootName string, indent bool) (string, error) {
	var sb strings.Builder
	err := writeXMLElement(&sb, rootName, data, 0, indent)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func writeXMLElement(sb *strings.Builder, name string, value interface{}, depth int, indent bool) error {
	indentStr := ""
	newline := ""
	if indent {
		indentStr = strings.Repeat("  ", depth)
		newline = "\n"
	}

	switch v := value.(type) {
	case map[string]interface{}:
		sb.WriteString(indentStr + "<" + name)

		// Write attributes first
		var childContent []struct {
			key   string
			value interface{}
		}
		for key, val := range v {
			if strings.HasPrefix(key, "@") {
				sb.WriteString(fmt.Sprintf(` %s="%v"`, strings.TrimPrefix(key, "@"), val))
			} else {
				childContent = append(childContent, struct {
					key   string
					value interface{}
				}{key, val})
			}
		}

		if len(childContent) == 0 {
			sb.WriteString("/>" + newline)
		} else {
			sb.WriteString(">" + newline)
			for _, child := range childContent {
				if err := writeXMLElement(sb, child.key, child.value, depth+1, indent); err != nil {
					return err
				}
			}
			sb.WriteString(indentStr + "</" + name + ">" + newline)
		}

	case []interface{}:
		for _, item := range v {
			if err := writeXMLElement(sb, name, item, depth, indent); err != nil {
				return err
			}
		}

	default:
		sb.WriteString(fmt.Sprintf("%s<%s>%v</%s>%s", indentStr, name, v, name, newline))
	}

	return nil
}

func navigateXPath(data map[string]interface{}, xpath string) interface{} {
	// Simple XPath: /root/child/subchild
	path := strings.TrimPrefix(xpath, "/")
	parts := strings.Split(path, "/")

	current := interface{}(data)
	for _, part := range parts {
		if part == "" {
			continue
		}

		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return current
}

// Note: XMLNode is registered in logic/init.go

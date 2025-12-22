package logic

import (
	"context"
	"regexp"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// HTMLExtractNode extracts data from HTML content
type HTMLExtractNode struct{}

func (n *HTMLExtractNode) Type() string {
	return "logic.html_extract"
}

func (n *HTMLExtractNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	html := core.GetString(config, "html", "")
	if html == "" {
		if h, ok := input["html"].(string); ok {
			html = h
		} else if h, ok := input["body"].(string); ok {
			html = h
		} else if h, ok := input["content"].(string); ok {
			html = h
		}
	}

	operation := core.GetString(config, "operation", "text")

	switch operation {
	case "text":
		return n.extractText(html, config)
	case "attribute":
		return n.extractAttribute(html, config)
	case "links":
		return n.extractLinks(html, config)
	case "images":
		return n.extractImages(html, config)
	case "tables":
		return n.extractTables(html, config)
	case "meta":
		return n.extractMeta(html)
	case "css":
		return n.extractByCSS(html, config)
	case "xpath":
		return n.extractByXPath(html, config)
	default:
		return n.extractText(html, config)
	}
}

func (n *HTMLExtractNode) extractText(html string, config map[string]interface{}) (map[string]interface{}, error) {
	selector := core.GetString(config, "selector", "")

	var text string
	if selector != "" {
		text = extractBySelector(html, selector)
	} else {
		text = stripTags(html)
	}

	text = strings.TrimSpace(text)
	text = normalizeWhitespace(text)

	return map[string]interface{}{
		"text":   text,
		"length": len(text),
	}, nil
}

func (n *HTMLExtractNode) extractAttribute(html string, config map[string]interface{}) (map[string]interface{}, error) {
	selector := core.GetString(config, "selector", "")
	attribute := core.GetString(config, "attribute", "href")

	values := extractAttributes(html, selector, attribute)

	return map[string]interface{}{
		"values": values,
		"count":  len(values),
		"first":  firstOrEmpty(values),
	}, nil
}

func (n *HTMLExtractNode) extractLinks(html string, config map[string]interface{}) (map[string]interface{}, error) {
	baseURL := core.GetString(config, "baseUrl", "")
	onlyExternal := core.GetBool(config, "onlyExternal", false)

	linkRegex := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)

	var links []map[string]interface{}
	for _, match := range matches {
		if len(match) >= 3 {
			href := match[1]
			text := stripTags(match[2])

			// Make absolute URL
			if baseURL != "" && !strings.HasPrefix(href, "http") {
				if strings.HasPrefix(href, "/") {
					href = baseURL + href
				} else {
					href = baseURL + "/" + href
				}
			}

			// Filter external links
			if onlyExternal && baseURL != "" {
				if strings.HasPrefix(href, baseURL) {
					continue
				}
			}

			links = append(links, map[string]interface{}{
				"href": href,
				"text": strings.TrimSpace(text),
			})
		}
	}

	return map[string]interface{}{
		"links": links,
		"count": len(links),
	}, nil
}

func (n *HTMLExtractNode) extractImages(html string, config map[string]interface{}) (map[string]interface{}, error) {
	baseURL := core.GetString(config, "baseUrl", "")

	imgRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	altRegex := regexp.MustCompile(`alt=["']([^"']*)["']`)

	matches := imgRegex.FindAllStringSubmatch(html, -1)

	var images []map[string]interface{}
	for _, match := range matches {
		if len(match) >= 2 {
			src := match[1]

			// Make absolute URL
			if baseURL != "" && !strings.HasPrefix(src, "http") {
				if strings.HasPrefix(src, "/") {
					src = baseURL + src
				} else {
					src = baseURL + "/" + src
				}
			}

			alt := ""
			if altMatch := altRegex.FindStringSubmatch(match[0]); len(altMatch) >= 2 {
				alt = altMatch[1]
			}

			images = append(images, map[string]interface{}{
				"src": src,
				"alt": alt,
			})
		}
	}

	return map[string]interface{}{
		"images": images,
		"count":  len(images),
	}, nil
}

func (n *HTMLExtractNode) extractTables(html string, config map[string]interface{}) (map[string]interface{}, error) {
	tableIndex := core.GetInt(config, "tableIndex", 0)
	hasHeader := core.GetBool(config, "hasHeader", true)

	tableRegex := regexp.MustCompile(`(?s)<table[^>]*>(.*?)</table>`)
	tables := tableRegex.FindAllStringSubmatch(html, -1)

	if tableIndex >= len(tables) {
		return map[string]interface{}{
			"rows":    []interface{}{},
			"headers": []string{},
			"count":   0,
		}, nil
	}

	tableHTML := tables[tableIndex][1]

	// Extract rows
	rowRegex := regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	rows := rowRegex.FindAllStringSubmatch(tableHTML, -1)

	var headers []string
	var data []map[string]interface{}

	for i, row := range rows {
		cellRegex := regexp.MustCompile(`(?s)<t[dh][^>]*>(.*?)</t[dh]>`)
		cells := cellRegex.FindAllStringSubmatch(row[1], -1)

		if i == 0 && hasHeader {
			for _, cell := range cells {
				headers = append(headers, strings.TrimSpace(stripTags(cell[1])))
			}
		} else {
			rowData := make(map[string]interface{})
			for j, cell := range cells {
				value := strings.TrimSpace(stripTags(cell[1]))
				if j < len(headers) {
					rowData[headers[j]] = value
				} else {
					rowData[string(rune('A'+j))] = value
				}
			}
			data = append(data, rowData)
		}
	}

	return map[string]interface{}{
		"rows":    data,
		"headers": headers,
		"count":   len(data),
	}, nil
}

func (n *HTMLExtractNode) extractMeta(html string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Title
	titleRegex := regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	if match := titleRegex.FindStringSubmatch(html); len(match) >= 2 {
		result["title"] = strings.TrimSpace(match[1])
	}

	// Meta tags
	metaRegex := regexp.MustCompile(`<meta[^>]+>`)
	metaTags := metaRegex.FindAllString(html, -1)

	meta := make(map[string]string)
	for _, tag := range metaTags {
		nameRegex := regexp.MustCompile(`(?:name|property)=["']([^"']+)["']`)
		contentRegex := regexp.MustCompile(`content=["']([^"']*)["']`)

		nameMatch := nameRegex.FindStringSubmatch(tag)
		contentMatch := contentRegex.FindStringSubmatch(tag)

		if len(nameMatch) >= 2 && len(contentMatch) >= 2 {
			meta[nameMatch[1]] = contentMatch[1]
		}
	}
	result["meta"] = meta

	// Open Graph
	og := make(map[string]string)
	for key, value := range meta {
		if strings.HasPrefix(key, "og:") {
			og[strings.TrimPrefix(key, "og:")] = value
		}
	}
	result["openGraph"] = og

	// Canonical URL
	canonicalRegex := regexp.MustCompile(`<link[^>]+rel=["']canonical["'][^>]+href=["']([^"']+)["']`)
	if match := canonicalRegex.FindStringSubmatch(html); len(match) >= 2 {
		result["canonical"] = match[1]
	}

	return result, nil
}

func (n *HTMLExtractNode) extractByCSS(html string, config map[string]interface{}) (map[string]interface{}, error) {
	selector := core.GetString(config, "selector", "")
	if selector == "" {
		return map[string]interface{}{"elements": []interface{}{}, "count": 0}, nil
	}

	elements := extractBySelector(html, selector)

	return map[string]interface{}{
		"elements": elements,
		"text":     stripTags(elements),
		"count":    1,
	}, nil
}

func (n *HTMLExtractNode) extractByXPath(html string, config map[string]interface{}) (map[string]interface{}, error) {
	// Basic XPath support (simplified)
	xpath := core.GetString(config, "xpath", "")
	if xpath == "" {
		return map[string]interface{}{"elements": []interface{}{}, "count": 0}, nil
	}

	// Convert simple XPath to regex pattern
	// This is a simplified implementation
	var pattern string
	if strings.HasPrefix(xpath, "//") {
		tag := strings.TrimPrefix(xpath, "//")
		if idx := strings.Index(tag, "/"); idx != -1 {
			tag = tag[:idx]
		}
		if idx := strings.Index(tag, "["); idx != -1 {
			tag = tag[:idx]
		}
		pattern = `(?s)<` + tag + `[^>]*>(.*?)</` + tag + `>`
	}

	if pattern == "" {
		return map[string]interface{}{"elements": []interface{}{}, "count": 0}, nil
	}

	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(html, -1)

	var elements []string
	for _, match := range matches {
		if len(match) >= 2 {
			elements = append(elements, match[1])
		}
	}

	return map[string]interface{}{
		"elements": elements,
		"count":    len(elements),
	}, nil
}

// Helper functions

func stripTags(html string) string {
	// Remove script tags with their content
	scriptRegex := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")

	// Remove style tags with their content
	styleRegex := regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")

	// Remove all HTML tags
	tagRegex := regexp.MustCompile(`<[^>]+>`)
	return tagRegex.ReplaceAllString(html, "")
}

func normalizeWhitespace(s string) string {
	wsRegex := regexp.MustCompile(`\s+`)
	return wsRegex.ReplaceAllString(s, " ")
}

func extractBySelector(html, selector string) string {
	// Simple selector support (tag, .class, #id)
	var pattern string

	if strings.HasPrefix(selector, "#") {
		id := strings.TrimPrefix(selector, "#")
		pattern = `(?s)<[^>]+id=["']` + id + `["'][^>]*>(.*?)</[^>]+>`
	} else if strings.HasPrefix(selector, ".") {
		class := strings.TrimPrefix(selector, ".")
		pattern = `(?s)<[^>]+class=["'][^"']*` + class + `[^"']*["'][^>]*>(.*?)</[^>]+>`
	} else {
		pattern = `(?s)<` + selector + `[^>]*>(.*?)</` + selector + `>`
	}

	regex := regexp.MustCompile(pattern)
	if match := regex.FindStringSubmatch(html); len(match) >= 2 {
		return match[1]
	}
	return ""
}

func extractAttributes(html, selector, attribute string) []string {
	var pattern string
	if selector == "" {
		pattern = `<[^>]+` + attribute + `=["']([^"']+)["'][^>]*>`
	} else {
		pattern = `<` + selector + `[^>]+` + attribute + `=["']([^"']+)["'][^>]*>`
	}

	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(html, -1)

	var values []string
	for _, match := range matches {
		if len(match) >= 2 {
			values = append(values, match[1])
		}
	}
	return values
}

func firstOrEmpty(arr []string) string {
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}

// Note: HTMLExtractNode is registered in logic/init.go

package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"plugin"
	"strings"
)

// NodeExecutor is the interface for all node executors
type NodeExecutor interface {
	Execute(config map[string]interface{}, input map[string]interface{}) (interface{}, error)
}

// LoadExecutor dynamically loads an executor
func LoadExecutor(executorClass string) (NodeExecutor, error) {
	// For built-in executors
	switch executorClass {
	case "httpRequest":
		return &HttpRequestExecutor{}, nil
	case "filter":
		return &FilterExecutor{}, nil
	case "transform":
		return &TransformExecutor{}, nil
	}

	// For plugins (dynamically loaded executors)
	if strings.HasPrefix(executorClass, "plugin:") {
		pluginPath := strings.TrimPrefix(executorClass, "plugin:")
		return loadPluginExecutor(pluginPath)
	}

	return nil, fmt.Errorf("unknown executor class: %s", executorClass)
}

// loadPluginExecutor loads an executor from a Go plugin
func loadPluginExecutor(pluginPath string) (NodeExecutor, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}

	symbol, err := p.Lookup("NewExecutor")
	if err != nil {
		return nil, err
	}

	newExecutorFunc, ok := symbol.(func() NodeExecutor)
	if !ok {
		return nil, fmt.Errorf("plugin does not provide a valid NewExecutor function")
	}

	return newExecutorFunc(), nil
}

// HttpRequestExecutor executes HTTP requests
type HttpRequestExecutor struct{}

func (e *HttpRequestExecutor) Execute(config map[string]interface{}, input map[string]interface{}) (interface{}, error) {
	// Get URL from configuration
	url, ok := config["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required in config")
	}

	// Get method from configuration or use default
	method, _ := config["method"].(string)
	if method == "" {
		method = "GET"
	}

	// Get headers from configuration
	headers := make(map[string]string)
	if headersConfig, ok := config["headers"].(map[string]interface{}); ok {
		for key, value := range headersConfig {
			if strValue, ok := value.(string); ok {
				headers[key] = strValue
			}
		}
	}

	// Replace template placeholders in the URL
	if strings.Contains(url, "{{") && strings.Contains(url, "}}") {
		for key, value := range input {
			placeholder := "{{" + key + "}}"
			if strings.Contains(url, placeholder) {
				url = strings.Replace(url, placeholder, fmt.Sprintf("%v", value), -1)
			}
		}
	}

	// Create HTTP client
	client := &http.Client{}

	// Prepare HTTP request
	var req *http.Request
	var err error

	if method == "GET" || method == "DELETE" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		// Get JSON data for POST/PUT from configuration
		var jsonData []byte
		if data, ok := config["json_data"]; ok {
			jsonData, err = json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal json data: %v", err)
			}
		}

		req, err = http.NewRequest(method, url, strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Try to parse the response as JSON
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// If not JSON, return as text
		result = map[string]interface{}{
			"text": string(body),
		}
	}

	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"data":        result,
	}, nil
}

// FilterExecutor filters data based on conditions
type FilterExecutor struct{}

func (e *FilterExecutor) Execute(config map[string]interface{}, input map[string]interface{}) (interface{}, error) {
	// Filter configuration
	filterField, _ := config["field"].(string)
	filterOperator, _ := config["operator"].(string)
	if filterOperator == "" {
		filterOperator = "equals"
	}
	filterValue := config["value"]

	// Read input data
	var items []interface{}

	// If there's only one input, use its value
	if len(input) == 1 {
		for _, v := range input {
			// Make sure we have a list
			switch val := v.(type) {
			case []interface{}:
				items = val
			default:
				// If it's not a list, create one with a single element
				items = []interface{}{val}
			}
			break
		}
	} else {
		// Use the default value "input" if available
		if inputArray, ok := input["input"].([]interface{}); ok {
			items = inputArray
		}
	}

	// Filter the elements
	var filtered []interface{}

	for _, item := range items {
		// Get the value from the item (also supports nested paths)
		itemValue := e.getNestedValue(item, filterField)

		// Apply the filter
		if e.compareValues(itemValue, filterValue, filterOperator) {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// getNestedValue gets a nested value from an object
func (e *FilterExecutor) getNestedValue(item interface{}, fieldPath string) interface{} {
	if fieldPath == "" {
		return item
	}

	parts := strings.Split(fieldPath, ".")
	current := item

	for _, part := range parts {
		if mapItem, ok := current.(map[string]interface{}); ok {
			if value, exists := mapItem[part]; exists {
				current = value
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return current
}

// compareValues compares two values with the specified operator
func (e *FilterExecutor) compareValues(value1, value2 interface{}, operator string) bool {
	switch operator {
	case "equals":
		return fmt.Sprintf("%v", value1) == fmt.Sprintf("%v", value2)
	case "not_equals":
		return fmt.Sprintf("%v", value1) != fmt.Sprintf("%v", value2)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", value1), fmt.Sprintf("%v", value2))
	case "greater_than":
		// For numeric comparisons, float conversions would need to happen here
		return false // Simplified implementation
	case "less_than":
		return false // Simplified implementation
	default:
		return false
	}
}

// TransformExecutor transforms data based on a mapping template
type TransformExecutor struct{}

func (e *TransformExecutor) Execute(config map[string]interface{}, input map[string]interface{}) (interface{}, error) {
	// Mapping-Template aus der Konfiguration holen
	mapping, ok := config["mapping"]
	if !ok {
		return nil, fmt.Errorf("mapping is required in config")
	}

	// Eingangsdaten auslesen
	var items []interface{}

	// Wenn es nur einen Eingang gibt, verwende dessen Wert
	if len(input) == 1 {
		for _, v := range input {
			// Stelle sicher, dass wir eine Liste haben
			switch val := v.(type) {
			case []interface{}:
				items = val
			default:
				// Wenn es keine Liste ist, erstelle eine mit einem Element
				items = []interface{}{val}
			}
			break
		}
	} else {
		// Verwende den Standardwert "input", falls vorhanden
		if inputArray, ok := input["input"].([]interface{}); ok {
			items = inputArray
		}
	}

	// Wende das Mapping auf jedes Element an
	var result []interface{}

	for _, item := range items {
		transformedItem := e.applyMapping(item, mapping)
		result = append(result, transformedItem)
	}

	return result, nil
}

// applyMapping wendet ein Mapping-Template auf ein Item an
func (e *TransformExecutor) applyMapping(item, mapping interface{}) interface{} {
	switch m := mapping.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range m {
			switch v := value.(type) {
			case string:
				// Prüfe auf Template-Ausdrücke wie "{{ data.name }}"
				if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
					path := strings.TrimSpace(v[2 : len(v)-2])
					result[key] = e.getNestedValue(item, path)
				} else {
					result[key] = v
				}
			case map[string]interface{}, []interface{}:
				result[key] = e.applyMapping(item, v)
			default:
				result[key] = v
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(m))
		for i, value := range m {
			result[i] = e.applyMapping(item, value)
		}
		return result
	default:
		return mapping
	}
}

// getNestedValue holt einen verschachtelten Wert aus einem Objekt
func (e *TransformExecutor) getNestedValue(item interface{}, fieldPath string) interface{} {
	if fieldPath == "" {
		return item
	}

	parts := strings.Split(fieldPath, ".")
	current := item

	for _, part := range parts {
		if mapItem, ok := current.(map[string]interface{}); ok {
			if value, exists := mapItem[part]; exists {
				current = value
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return current
}

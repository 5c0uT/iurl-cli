package query

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Apply(data []byte, expr string) ([]byte, error) {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	result, err := evaluate(parsed, expr)
	if err != nil {
		return nil, err
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return out, nil
}

func evaluate(data interface{}, expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return data, nil
	}

	if strings.HasPrefix(expr, ".") {
		return evalDot(data, expr)
	}

	if strings.HasPrefix(expr, "[") {
		return evalIndex(data, expr)
	}

	if strings.HasPrefix(expr, "select(") {
		return evalSelect(data, expr)
	}

	if strings.HasPrefix(expr, "map(") {
		return evalMap(data, expr)
	}

	if expr == "length" {
		return evalLength(data)
	}

	if expr == "keys" {
		return evalKeys(data)
	}

	if expr == "values" {
		return evalValues(data)
	}

	if expr == "type" {
		return evalType(data)
	}

	if expr == "null" {
		return nil, nil
	}

	if expr == "true" {
		return true, nil
	}

	if expr == "false" {
		return false, nil
	}

	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return strings.Trim(expr, "\""), nil
	}

	if n, err := strconv.ParseFloat(expr, 64); err == nil {
		return n, nil
	}

	return data, nil
}

func evalDot(data interface{}, expr string) (interface{}, error) {
	expr = strings.TrimPrefix(expr, ".")
	if expr == "" {
		return data, nil
	}

	parts := strings.SplitN(expr, ".", 2)
	key := parts[0]

	switch v := data.(type) {
	case map[string]interface{}:
		if val, ok := v[key]; ok {
			if len(parts) > 1 {
				return evaluate(val, "."+parts[1])
			}
			return val, nil
		}
		return nil, fmt.Errorf("key not found: %s", key)
	case []interface{}:
		if idx, err := strconv.Atoi(key); err == nil && idx >= 0 && idx < len(v) {
			if len(parts) > 1 {
				return evaluate(v[idx], "."+parts[1])
			}
			return v[idx], nil
		}
		return nil, fmt.Errorf("index out of range: %s", key)
	}

	return nil, fmt.Errorf("cannot access .%s on %T", key, data)
}

func evalIndex(data interface{}, expr string) (interface{}, error) {
	expr = strings.Trim(expr, "[]")
	if idx, err := strconv.Atoi(expr); err == nil {
		if arr, ok := data.([]interface{}); ok && idx >= 0 && idx < len(arr) {
			return arr[idx], nil
		}
		return nil, fmt.Errorf("index out of range: %d", idx)
	}

	return nil, fmt.Errorf("invalid index: %s", expr)
}

func evalSelect(data interface{}, expr string) (interface{}, error) {
	inner := strings.TrimPrefix(expr, "select(")
	inner = strings.TrimSuffix(inner, ")")

	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("select requires array, got %T", data)
	}

	var result []interface{}
	for _, item := range arr {
		matched, err := matchCondition(item, inner)
		if err != nil {
			return nil, err
		}
		if matched {
			result = append(result, item)
		}
	}

	return result, nil
}

func matchCondition(item interface{}, cond string) (bool, error) {
	cond = strings.TrimSpace(cond)

	if strings.Contains(cond, "==") {
		parts := strings.SplitN(cond, "==", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"'")

		val2, err := evaluate(item, key)
		if err != nil {
			return false, nil
		}

		if s, ok := val2.(string); ok {
			return s == val, nil
		}
		if f, ok := val2.(float64); ok {
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				return f == fv, nil
			}
		}
		if b, ok := val2.(bool); ok {
			if val == "true" {
				return b == true, nil
			}
			if val == "false" {
				return b == false, nil
			}
		}
		return false, nil
	}

	if strings.Contains(cond, "!=") {
		parts := strings.SplitN(cond, "!=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"'")

		val2, err := evaluate(item, key)
		if err != nil {
			return false, nil
		}

		if s, ok := val2.(string); ok {
			return s != val, nil
		}
		return false, nil
	}

	if strings.Contains(cond, ">") && !strings.Contains(cond, ">=") {
		parts := strings.SplitN(cond, ">", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		val2, err := evaluate(item, key)
		if err != nil {
			return false, nil
		}

		if f, ok := val2.(float64); ok {
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				return f > fv, nil
			}
		}
		return false, nil
	}

	return false, nil
}

func evalMap(data interface{}, expr string) (interface{}, error) {
	inner := strings.TrimPrefix(expr, "map(")
	inner = strings.TrimSuffix(inner, ")")

	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("map requires array, got %T", data)
	}

	var result []interface{}
	for _, item := range arr {
		val, err := evaluate(item, inner)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}

	return result, nil
}

func evalLength(data interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return float64(v.Len()), nil
	case reflect.Map:
		return float64(v.Len()), nil
	case reflect.String:
		return float64(v.Len()), nil
	}
	return float64(1), nil
}

func evalKeys(data interface{}) (interface{}, error) {
	if m, ok := data.(map[string]interface{}); ok {
		keys := make([]interface{}, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys, nil
	}
	return nil, fmt.Errorf("keys requires object, got %T", data)
}

func evalValues(data interface{}) (interface{}, error) {
	if m, ok := data.(map[string]interface{}); ok {
		vals := make([]interface{}, 0, len(m))
		for _, v := range m {
			vals = append(vals, v)
		}
		return vals, nil
	}
	return nil, fmt.Errorf("values requires object, got %T", data)
}

func evalType(data interface{}) (interface{}, error) {
	if data == nil {
		return "null", nil
	}
	switch data.(type) {
	case string:
		return "string", nil
	case float64:
		return "number", nil
	case bool:
		return "boolean", nil
	case map[string]interface{}:
		return "object", nil
	case []interface{}:
		return "array", nil
	}
	return "unknown", nil
}
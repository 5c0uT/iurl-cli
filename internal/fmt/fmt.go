package fmt

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

const (
	colorReset = "\033[0m"
	colorKey   = "\033[1;34m"
	colorStr   = "\033[32m"
	colorNum   = "\033[33m"
	colorBool  = "\033[35m"
	colorNull  = "\033[36m"
)

func PrettyPrintJSON(reader io.Reader, writer io.Writer) error {
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)

	var data interface{}
	decoder := json.NewDecoder(tee)
	if err := decoder.Decode(&data); err != nil {
		io.Copy(writer, &buf)
		return err
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		io.Copy(writer, &buf)
		return err
	}

	io.WriteString(writer, ColorizeJSON(prettyJSON))
	return nil
}

func CopyRaw(reader io.Reader, writer io.Writer) error {
	_, err := io.Copy(writer, reader)
	return err
}

func ColorizeJSON(data []byte) string {
	var out strings.Builder
	i := 0
	n := len(data)

	for i < n {
		ch := data[i]

		switch {
		case ch == '"':
			start := i
			i++
			for i < n {
				if data[i] == '\\' {
					i += 2
					continue
				}
				if data[i] == '"' {
					i++
					break
				}
				i++
			}
			raw := string(data[start:i])
			if isKeyAt(data, i) {
				out.WriteString(colorKey)
				out.WriteString(raw)
				out.WriteString(colorReset)
				out.WriteString(": ")
				i += 2
			} else {
				out.WriteString(colorStr)
				out.WriteString(raw)
				out.WriteString(colorReset)
			}
		case ch == '-' || (ch >= '0' && ch <= '9'):
			start := i
			if ch == '-' {
				i++
			}
			for i < n && data[i] >= '0' && data[i] <= '9' {
				i++
			}
			if i < n && data[i] == '.' {
				i++
				for i < n && data[i] >= '0' && data[i] <= '9' {
					i++
				}
			}
			if i < n && (data[i] == 'e' || data[i] == 'E') {
				i++
				if i < n && (data[i] == '+' || data[i] == '-') {
					i++
				}
				for i < n && data[i] >= '0' && data[i] <= '9' {
					i++
				}
			}
			out.WriteString(colorNum)
			out.WriteString(string(data[start:i]))
			out.WriteString(colorReset)
		case ch == 't':
			out.WriteString(colorBool)
			out.WriteString("true")
			out.WriteString(colorReset)
			i += 4
		case ch == 'f':
			out.WriteString(colorBool)
			out.WriteString("false")
			out.WriteString(colorReset)
			i += 5
		case ch == 'n':
			out.WriteString(colorNull)
			out.WriteString("null")
			out.WriteString(colorNull)
			i += 4
		default:
			out.WriteByte(ch)
			i++
		}
	}

	return out.String()
}

func isKeyAt(data []byte, pos int) bool {
	for pos < len(data) {
		ch := data[pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			pos++
			continue
		}
		return ch == ':'
	}
	return false
}
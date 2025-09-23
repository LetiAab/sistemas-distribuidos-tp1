//go:build ignore
// +build ignore

package main

import (
	"path/filepath"
	"sistemas-distribuidos-tp1/internal/common"
	"strings"
	"unicode"
)

var Map func(string, string) []common.KeyValue
var Reduce func(string, []string) string

func init() {
	Map = func(filename string, content string) []common.KeyValue {
		var kvs []common.KeyValue

		ff := func(r rune) bool { return !unicode.IsLetter(r) }
		words := strings.FieldsFunc(content, ff)

		baseFilename := filepath.Base(filename)

		// para ii, emitimos (palabra, archivo_donde_aparece)
		for _, word := range words {
			normalizedWord := strings.ToLower(word)
			kvs = append(kvs, common.KeyValue{Key: normalizedWord, Value: baseFilename})
		}

		return kvs
	}

	Reduce = func(key string, values []string) string {

		// Eliminar duplicados usando un map como set
		uniqueFiles := make(map[string]bool)
		for _, file := range values {
			uniqueFiles[file] = true
		}

		// Convertir a slice ordenado
		var fileList []string
		for file := range uniqueFiles {
			fileList = append(fileList, file)
		}

		// Unir con comas
		return strings.Join(fileList, ",")
	}
}

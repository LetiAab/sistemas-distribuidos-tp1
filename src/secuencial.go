package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
)

type KeyValue struct {
	Key   string
	Value int
}

func Map(filename string, content string) []KeyValue {
	var kvs []KeyValue
	ff := func(r rune) bool { return !unicode.IsLetter(r) } // Filtra caracteres no alfabéticos
	words := strings.FieldsFunc(content, ff)                // Separa el contenido en palabras

	for _, word := range words {
		normalizedWord := strings.ToLower(word) // Convierte a minúsculas
		kvs = append(kvs, KeyValue{normalizedWord, 1})
	}
	return kvs
}

func Reduce(key string, values []int) int {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum
}

func main() {
	inputFiles := os.Args[1:]

	var intermediate []KeyValue
	for _, filename := range inputFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Printf("Error leyendo %s: %v\n", filename, err)
			continue
		}
		kvs := Map(filename, string(content))
		intermediate = append(intermediate, kvs...)
	}

	groups := make(map[string][]int)
	for _, kv := range intermediate { // Agrupamos tal que quede - (Hello, [1, 1])
		groups[kv.Key] = append(groups[kv.Key], kv.Value)
	}

	var results []KeyValue
	for key, values := range groups {
		count := Reduce(key, values)
		results = append(results, KeyValue{key, count})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})

	outputFile := "mr-out-0"
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creando archivo de salida: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, kv := range results {
		fmt.Fprintf(writer, "%v %v\n", kv.Key, kv.Value)
	}
	writer.Flush()
}

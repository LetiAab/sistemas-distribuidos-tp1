package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sistemas-distribuidos-tp1/common"
	"strconv"
	"strings"
	"unicode"
)

func Map(filename string, content string) []common.KeyValue {
	var kvs []common.KeyValue

	ff := func(r rune) bool { return !unicode.IsLetter(r) } // Filtra caracteres no alfabéticos
	words := strings.FieldsFunc(content, ff)                // Separa el contenido en palabras

	for _, word := range words {
		normalizedWord := strings.ToLower(word) // Convierte a minúsculas
		kvs = append(kvs, common.KeyValue{Key: normalizedWord, Value: "1"})
	}

	return kvs
}

func Reduce(key string, values []string) string {
	sum := 0
	for _, v := range values {
		num, err := strconv.Atoi(v)
		if err != nil {
			fmt.Println("Hubo un error en: ", err)
			continue
		}
		sum += num
	}

	return strconv.Itoa(sum)
}

func main() {
	inputFiles := os.Args[1:]

	var intermediate []common.KeyValue
	for _, filename := range inputFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Printf("Error leyendo %s: %v\n", filename, err)
			continue
		}
		kvs := Map(filename, string(content))
		intermediate = append(intermediate, kvs...)
	}

	groups := make(map[string][]string)
	for _, kv := range intermediate { // Agrupamos tal que quede - (Hello, [1, 1])
		groups[kv.Key] = append(groups[kv.Key], kv.Value)
	}

	var results []common.KeyValue
	for key, values := range groups {
		count := Reduce(key, values)
		results = append(results, common.KeyValue{Key: key, Value: count})
	}

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

package main

import (
	"log"
	"math/rand"
	"sistemas-distribuidos-tp1/common"
	"strconv"
	"strings"
	"unicode"
)

var Map func(string, string) []common.KeyValue
var Reduce func(string, []string) string

func init() {
	Map = func(filename string, content string) []common.KeyValue {
		randomValue := rand.Float32()
		log.Printf("Map - Random value: %f", randomValue)

		if randomValue < 0.05 {
			panic("Fallo simulado en Map")
		}

		var kvs []common.KeyValue
		ff := func(r rune) bool { return !unicode.IsLetter(r) }
		words := strings.FieldsFunc(content, ff)

		for _, word := range words {
			normalizedWord := strings.ToLower(word)
			kvs = append(kvs, common.KeyValue{normalizedWord, "1"})
		}

		return kvs
	}

	Reduce = func(key string, values []string) string {
		randomValue := rand.Float32()

		if randomValue < 0.000005 {
			panic("Fallo simulado en Reduce")
		}

		sum := 0
		for _, v := range values {
			num, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			sum += num
		}
		return strconv.Itoa(sum)
	}
}

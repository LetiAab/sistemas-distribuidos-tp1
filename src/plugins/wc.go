package main

import (
    "strconv"
    "strings"
    "unicode"
	"fmt"
)

type KeyValue struct {
    Key   string
    Value string
}

// Defino así las funcs porque Go carga Plugins dinamicamente, 
// osea q busca variables públicas. 
var Map func(string, string) []KeyValue 
var Reduce func(string, []string) string 


func init() {

	Map = func(filename string, content string) []KeyValue{
		var kvs []KeyValue

		// TODO: Esto no separa entre min y mayus creo
		ff := func(r rune) bool { return !unicode.IsLetter(r) }
		words := strings.FieldsFunc(content, ff)
	
		for _, word := range words{
			kvs = append(kvs, KeyValue{word, "1"})
		}
		
		return kvs
	}

	Reduce = func(key string, values []string) string{

		sum := 0
		for _, v := range values{
			num, err := strconv.Atoi(v)
			if err != nil{
				fmt.Println("Hubo un error en: ", err)
				continue
			}
			sum += num
		}

		return strconv.Itoa(sum)
	}

}
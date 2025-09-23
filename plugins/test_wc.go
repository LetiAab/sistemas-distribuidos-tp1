package main

import (
	"fmt"
	"log"
	"plugin"
	"sistemas-distribuidos-tp1/internal/common"
)

func main() {
	// Cargar el plugin
	p, err := plugin.Open("wc.so")
	if err != nil {
		log.Fatal(err)
	}

	// Obtener Map
	mapSymbol, _ := p.Lookup("Map")
	mapFunc := mapSymbol.(*func(string, string) []common.KeyValue)

	// PRUEBA 1: Map básico
	fmt.Println("=== Test Map ===")
	testText := "hello world hello"
	results := (*mapFunc)("test.txt", testText)

	for _, kv := range results {
		fmt.Printf("'%s' -> '%s'\n", kv.Key, kv.Value)
	}
	// Debería mostrar:
	// 'hello' -> '1'
	// 'world' -> '1'
	// 'hello' -> '1'

	// PRUEBA 2: Reduce
	fmt.Println("\n=== Test Reduce ===")
	reduceSymbol, _ := p.Lookup("Reduce")
	reduceFunc := reduceSymbol.(*func(string, []string) string)

	count := (*reduceFunc)("hello", []string{"1", "1", "1"})
	fmt.Printf("Count for 'hello': %s (esperado: 3)\n", count)

	// PRUEBA 3: Caso "Completo"
	fmt.Println("\n=== Test Completo ===")
	complexText := "La vida antes de la muerte. La fuerza antes de la debilidad. El viaje antes del destino."
	mapResults := (*mapFunc)("test.txt", complexText)

	// Agrupar manualmente para testing
	groups := make(map[string][]string)
	for _, kv := range mapResults {
		groups[kv.Key] = append(groups[kv.Key], kv.Value)
	}

	// Reducir cada grupo
	for word, values := range groups {
		count := (*reduceFunc)(word, values)
		fmt.Printf("%-10s: %s\n", word, count)
	}
}

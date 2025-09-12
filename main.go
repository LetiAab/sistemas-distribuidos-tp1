package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type KeyValue struct {
	key   string
	value int
}

// Remove punctuation and convert to lowercase
func sanitize(word string) string {
	return strings.ToLower(strings.TrimFunc(word, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}))
}

// for word count the key would be the file split id, maybe the
// file path or index offset in a file, and the value would be
// the text content of that split
func Map(key string, value string) []KeyValue {
	var acc []KeyValue

	words := strings.Fields(value)

	// TODO: optimization -> use a map to count occurrences

	for _, word := range words {
		word = sanitize(word)
		acc = append(acc, KeyValue{word, 1})
	}

	return acc
}

//	The MapReduce library groups together all intermediate values associated with the same intermediate key I and passes them to the Reduce function.
//
// for word count, the key would be the word and the value would be a list of counts for that word.
func Reduce(key string, values []int) int {
	var sum = 0

	for _, value := range values {
		sum += value
	}

	return sum
}

// Counts the number of occurences of each word in `text`, updating the `occurences` map.
func main() {
	// Parse command line flags
	sequential := flag.Bool("S", false, "Run in sequential mode")
	distributed := flag.Bool("D", false, "Run in distributed mode")
	flag.Parse()

	// Check that exactly one mode is specified
	if *sequential && *distributed {
		fmt.Println("Error: Cannot specify both -S and -D flags")
		fmt.Println("Usage: go run main.go (-S | -D) <directory>")
		return
	}
	if !*sequential && !*distributed {
		fmt.Println("Error: Must specify either -S (sequential) or -D (distributed) mode")
		fmt.Println("Usage: go run main.go (-S | -D) <directory>")
		return
	}

	// Check if exactly one directory argument is provided
	if len(flag.Args()) != 1 {
		fmt.Println("Usage: go run main.go (-S | -D) <directory>")
		return
	}

	dirPath := flag.Args()[0]
	info, err := os.Stat(dirPath)
	if err != nil {
		fmt.Printf("Error accessing %s: %v\n", dirPath, err)
		return
	}

	if !info.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", dirPath)
		return
	}

	// Read directory and collect all file paths
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", dirPath, err)
		return
	}

	var filePaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := dirPath + string(os.PathSeparator) + entry.Name()
			filePaths = append(filePaths, filePath)
		}
	}

	if len(filePaths) == 0 {
		fmt.Printf("No files found in directory %s\n", dirPath)
		return
	}

	if *distributed {
		// fmt.Println("Running in distributed mode...")
		runDistributed(filePaths)
	} else {
		// fmt.Println("Running in sequential mode...")
		runSequential(filePaths)
	}
}

func runSequential(filePaths []string) {
	// MAP

	var intermediate []KeyValue

	for _, filePath := range filePaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		keyvalues := Map(filePath, string(data))
		intermediate = append(intermediate, keyvalues...)
	}

	// SHUFFLE

	var intermediate_grouped = make(map[string][]int)

	for _, kv := range intermediate {
		intermediate_grouped[kv.key] = append(intermediate_grouped[kv.key], kv.value)
	}

	// REDUCE

	var word_count = make(map[string]int)

	for key, values := range intermediate_grouped {
		count := Reduce(key, values)

		word_count[key] = count

		fmt.Printf("%s, %d\n", key, count)
	}
}

func runDistributed(filePaths []string) {
	fmt.Println("Distributed mode not yet implemented")
}

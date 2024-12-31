package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/AndreiLacatos/opc-engine/node-engine/serialization"
)

func main() {
	input := os.Args[1]
	content, err := os.ReadFile(input)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	jsonString := string(content)

	var structure serialization.OpcStructureModel

	err = json.Unmarshal([]byte(jsonString), &structure)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	fmt.Printf("%v\n", structure)
	r := structure.Root.ToDomain()
	fmt.Printf("%v\n", r)
}
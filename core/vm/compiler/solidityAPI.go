package compiler

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
)

var (
	excFlag = flag.String("exc", "", "Comma separated types to exclude from binding")
)

func CompileSourcefile(source string) ([]string, []string, error) {
	var (
		abis  []string
		bins  []string
		types []string
	)
	solc, err := NewCompiler("")
	if err != nil {
		return nil, nil, err
	}
	contracts, err := solc.Compile(string(source))

	exclude := make(map[string]bool)
	for _, kind := range strings.Split(*excFlag, ",") {
		exclude[strings.ToLower(kind)] = true
	}
	if err != nil {
		fmt.Printf("Failed to build Solidity contract: %v\n", err)
	}
	// Gather all non-excluded contract for binding
	for name, contract := range contracts {
		if exclude[strings.ToLower(name)] {
			continue
		}
		abi, _ := json.Marshal(contract.Info.AbiDefinition) // Flatten the compiler parse
		abis = append(abis, string(abi))
		bins = append(bins, contract.Code)
		types = append(types, name)
	}
	return abis, bins, nil
}

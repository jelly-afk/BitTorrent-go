package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
		 bencodedValue := os.Args[2]
		//
	 decoded, err := decodeBencode(bencodedValue)
		 if err != nil {
		 	fmt.Println(err)
		 	return
		 }
		//
		 jsonOutput, _ := json.Marshal(decoded)
		 fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func decodeBencode(bencodedString string) (interface{}, error) {
    fval := rune(bencodedString[0])
	if unicode.IsDigit(fval) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if fval == 'i'{
        eIdx := strings.Index(bencodedString, "e")
        value, err := strconv.Atoi(bencodedString[1:eIdx])
        if err != nil {
            return "", err
        }
        return value, nil

    } else {
		return "", fmt.Errorf("Only strings are supported at the moment")
	}
}

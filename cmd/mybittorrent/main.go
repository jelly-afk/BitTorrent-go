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
	 _, decoded, err := decodeBencode(bencodedValue)
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

func decodeBencode(bencodedString string) (string, interface{}, error) {
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
			return "","", err
		}
        strEnd := firstColonIndex+1+length

        return bencodedString[strEnd:],bencodedString[firstColonIndex+1 : strEnd], nil
	} else if fval == 'i'{
        eIdx := strings.Index(bencodedString, "e")
        value, err := strconv.Atoi(bencodedString[1:eIdx])
        if err != nil {
            return "","",  err
        }
        return bencodedString[eIdx+1:],value, nil

    } else if fval == 'l'{
        res := []interface{}{}
        bencodedString = bencodedString[1:]
        for bencodedString[0] != byte('e'){
            var val interface{}
            var err error
            bencodedString, val, err = decodeBencode(bencodedString)
            if err != nil {
                return "", "", err
            }
            res = append(res, val)

        }
        return bencodedString[1:], res, nil
    } else if fval == 'd' {
        res := map[string]interface{}{}
        bencodedString = bencodedString[1:]
        for bencodedString[0] != byte('e'){
            var ikey interface{}
            var val interface{}
            var err error
            bencodedString, ikey, err = decodeBencode(bencodedString)
            if err != nil {
                return "", "", err
            }
            bencodedString, val, err = decodeBencode(bencodedString)
            if err != nil {
                return "", "", err
            }
        key := ikey.(string)
        res[key] = val
        }
        return bencodedString, res, nil

    }else {
		return "","", fmt.Errorf("Only strings are supported at the moment")
	}
}

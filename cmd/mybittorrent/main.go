package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
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
	} else if command == "info"{
        fileName := os.Args[2]
        _, infoMap, trackerUrl, err := parseTorrentFile(fileName)
        if err != nil {
            log.Fatal(err)
        }
        bencodedInfo, err := bencode(infoMap)

        if err != nil {
            log.Fatal(err)
        }
        hash := sha1.New()
        hash.Write([]byte(bencodedInfo))
        sha1Hash := hash.Sum(nil)
        hexHash := hex.EncodeToString(sha1Hash)
        fmt.Print("Tracker URL: ", trackerUrl)
        fmt.Printf("Piece Length: %d", infoMap["piece length"])
        pieceStr, ok := infoMap["pieces"].(string)
        if !ok {
            log.Fatal("invalid type")
        }
        fmt.Println("Piece Hashes: ")
        for len(pieceStr) > 0 {
            byteHash := []byte(pieceStr[:20])
            fmt.Printf("%x\n", byteHash)
            pieceStr = pieceStr[20:]
        } 
        fmt.Print("Info Hash: ", hexHash)
    }else if command == "peers" {
        fileName := os.Args[2]
        _, infoMap,trackerUrl,  err := parseTorrentFile(fileName)
        if err != nil {
            log.Fatal(err)
        }
        bencodedInfo, err := bencode(infoMap)
        if err != nil {
            log.Fatal(err)
        }
        urlStruc, err := url.Parse(trackerUrl)
        if err != nil {
            log.Fatal(err)
        }
        params := url.Values{}
        hash := sha1.New()
        hash.Write([]byte(bencodedInfo))
        sha1Hash := hash.Sum(nil)
        pieces := fmt.Sprintf("%v", infoMap["length"])
        params.Add("info_hash", string(sha1Hash))
        params.Add("peer_id", "qwertyuiopasdfghjklz")
        params.Add("port", "6881")
        params.Add("uploaded", "0")
        params.Add("downloaded", "0")
        params.Add("left", pieces)
        params.Add("compact", "1")
        urlStruc.RawQuery = params.Encode()
        res, err := http.Get(urlStruc.String())
        body, err := io.ReadAll(res.Body)
        if err != nil {
            log.Fatal(err)
        }
        _, decodedBody, err := decodeBencode(string(body))
        if err != nil {
            log.Fatal(err)
        }
        bodyMap, ok := decodedBody.(map[string]interface{})
        if !ok {
            log.Fatal("invalid type")
        }
        peers, ok := bodyMap["peers"].(string)
        if !ok {
            log.Fatal("invalid type")
        }
        peersByte := []byte(peers)
        for len(peersByte) > 0 {
            peer := peersByte[:6]
            port := binary.BigEndian.Uint16(peer[4:])
            fmt.Printf("%d.%d.%d.%d:%d\n", peer[0], peer[1] ,peer[2], peer[3], port)
            peersByte = peersByte[6:]
        }

    } else if command == "handshake"{
        fileName := os.Args[2]
        peerAddr := os.Args[3]
        _, infoMap,_ , err := parseTorrentFile(fileName)
        if err != nil {
            log.Fatal(fileName)
        }
        bencodedInfo, err  := bencode(infoMap)
        if err != nil {
            log.Fatal(err)
        }
        hash := sha1.New()
        hash.Write([]byte(bencodedInfo))
        sha1Hash := hash.Sum(nil)
        tcpConn, err := net.Dial("tcp", peerAddr)
        if err != nil {
            log.Fatal(err)
        }
        defer tcpConn.Close()
        buffer := make([]byte, 0, 68)
        buffer = append(buffer, 19)
        buffer = append(buffer, []byte("BitTorrent protocol")...)
        buffer = append(buffer, make([]byte, 8)...)
        buffer = append(buffer, sha1Hash...)
        buffer = append(buffer, make([]byte, 20)...)
        _, err = tcpConn.Write(buffer)
        if err != nil {
            log.Println("error while writing data in connection", err)
        }
        readBuffer := make([]byte, 68)
        _, err = tcpConn.Read(readBuffer)
        if err != nil {
            log.Fatal(err)
        }
        peerId := readBuffer[48:]
        fmt.Printf("Peer ID: %x\n", peerId)
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

func bencode(decoded interface{}) (string, error) {
    switch t := decoded.(type) {
    case int:
        return fmt.Sprintf("i%de", t), nil
    case string:
        return fmt.Sprintf("%d:%s", len(t), t), nil
    case []interface{}:
        var res []string
        for _, v := range t {
            s, err := bencode(v)
            if err != nil {
                log.Fatal(err)
            }
            res = append(res, s)
        }
        return fmt.Sprintf("l%se",strings.Join(res, "")), nil
    case map[string]interface{}:
        var keys []string
        dict := make(map[string]string)
        for k, v := range t {
            kB, err := bencode(k)
            if err != nil {
                log.Fatal(err)
            }
            vB, err := bencode(v)
            if err != nil {
                log.Fatal(err)
            }
            keys = append(keys, k)
            dict[kB] = vB
        }
        sort.Strings(keys)
        for i, k := range keys {
        kb,_  := bencode(k)
            keys[i] = kb+dict[kb]
    }
        return fmt.Sprintf("d%se", strings.Join(keys, "")), nil
    } 
    return "", errors.New("invalid type")
}

func parseTorrentFile (fileName string) ( map[string]interface{}, map[string]interface{},string,  error) {
    data, err  := os.ReadFile(fileName)
    if err != nil {
        return nil, nil,"",  err
    }
    _, decoded, err := decodeBencode(string(data))
    if err != nil {
        return nil, nil,"",  err
    }
    headerMap, ok := decoded.(map[string]interface{})
    if !ok {
        return nil, nil,"",  errors.New("invalid type")
    }
    infoMap, ok := headerMap["info"].(map[string]interface{})
    if !ok {
        return nil, nil,"",  errors.New("invalid type")
    }
    trackerUrl, ok := headerMap["announce"].(string)
    if !ok {
        log.Fatal(err)
    }
    return headerMap, infoMap,trackerUrl,  nil
}

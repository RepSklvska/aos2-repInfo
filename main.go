package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	Filename = "./aos2_039.rpl"
)

func getLastIndex(content []byte, target []byte) int {
	for i := len(content) - 1; i >= 0; i-- {
		if string(content[i-len(target):i]) == string(target) {
			return i - len(target)
		}
	}
	return -1
}

func getFirstIndex(content []byte, target []byte) int {
	for i, _ := range content {
		if string(content[i:i+len(target)]) == string(target) {
			return i
		}
	}
	return -1
}

func main() {
	content := func() []byte {
		if content, err := ioutil.ReadFile(os.Args[1]); err == nil {
			//fmt.Println(os.Args)
			return content
		} else {
			panic(err)
		}
		return nil
	}()
	content = content[getLastIndex(content, []byte{0x00, 0x40, 0x00, 0x40})+4:]
	//fmt.Println(string(content))
	
	player1, player2 := func() (string, string) {
		var i1 int
		var p1, p2 string
		for i, v := range content {
			if v == 0x00 {
				i1 = i
				p1 = string(content[:i])
				break
			}
		}
		for i, v := range content[i1:] {
			if v != 0x00 {
				i1 += i
				break
			}
		}
		for i, v := range content[i1:] {
			if v == 0x00 {
				p2 = string(content[i1 : i1+i])
				break
			}
		}
		return p1, p2
	}()
	
	s := []byte("---------------------------------------------------------------------------")
	content = content[:getLastIndex(content, s)]
	content = content[getLastIndex(content, s)+len(s):]
	
	info := func() map[string]string {
		var infoMap map[string]string
		infoMap = make(map[string]string)
		//x := strings.ReplaceAll(string(content), string([]byte{0x81, 0xA1}), "")
		//x = strings.ReplaceAll(string(content), ":", "")
		y := strings.Split(string(content), "\r\n")
		//fmt.Println(y, len(y))
		//fmt.Println(y[0]=="")
		//fmt.Println(strings.Split(y[6], string(rune(0x09))), len(strings.Split(y[6], string(rune(0x09)))))
		
		keys := []string{"GameMode", "1P-side", "2P-side", "Time"}
		for _, v := range y {
			for _, vv := range keys {
				if strings.Contains(v, vv) {
					z := strings.Split(v, string(rune(0x09)))
					if z[len(z)-1] == "[W]" {
						infoMap[vv] = z[len(z)-2]
						infoMap["Winner"] = vv[0:2]
						continue
					}
					infoMap[vv] = z[len(z)-1]
				}
			}
		}
		return infoMap
	}()
	//fmt.Println(info)
	
	fmt.Println("Replay Info:")
    fmt.Println("-------------------------")
	fmt.Println("Mode:\t" + info["GameMode"])
	fmt.Println("Time:\t" + info["Time"])
	fmt.Println("Winner:\t" + info["Winner"])
	fmt.Println("1P:")
	fmt.Println("\t" + info["1P-side"])
	fmt.Println("\t" + player1)
	fmt.Println("2P:")
	fmt.Println("\t" + info["2P-side"])
	fmt.Println("\t" + player2)
	fmt.Println("-------------------------")
	func() {
		fmt.Println("Press Ctrl-C to exit...")
		for {
			_ = 1
		}
	}()
}

//I just liked to write a program which I can't read.

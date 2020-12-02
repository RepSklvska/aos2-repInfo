package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func getLastIndex(content []byte, target []byte) int {
	if len(content) == 0 {
		return -1
	}
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

type Player struct {
	Character string
	Name      string
}

type Info struct {
	Filename  string
	Mode      string
	Time      string
	Winner    string
	Player1   Player
	Player2   Player
	IsBadFile bool
}

type Infos []Info

func (infos *Infos) Push(a ...Info) {
	*infos = append(*infos, a...)
}

func isRepFile(filename string) bool {
	return filepath.Ext(filename) == ".rpl" && !isDir(filename)
}

func isDir(filename string) bool {
	if file, err := os.Stat(filename); err != nil {
		return false
	} else {
		return file.IsDir()
	}
}

func readRepDir(dirname string) (Infos, error) {
	dirnameUnix := strings.ReplaceAll(dirname, "\\", "/")
	filenames := []string{}
	infos := Infos{}
	if files, err := ioutil.ReadDir(dirname); err != nil {
		return nil, err
	} else {
		for _, v := range files {
			filenames = append(filenames, dirnameUnix+"/"+v.Name())
		}
	}
	for _, filename := range filenames {
		if info := getReplayInfo(filename); !info.IsEmpty() {
			infos.Push(info)
		} else {
			infos.Push(Info{
				Filename:  filepath.Base(filename),
				IsBadFile: true,
			})
		}
	}
	return infos, nil
}

func getReplayInfo(filename string) Info {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	if !isRepFile(filename) {
		return Info{}
	}

	content := func() []byte {
		if content, err := ioutil.ReadFile(filename); err == nil {
			return content
		} else {
			panic(err)
		}
		return nil
	}()
	content = content[getLastIndex(content, []byte{0x00, 0x40, 0x00, 0x40})+4:]

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
		if strings.ContainsRune(p1, 0xfffd) {
			p1 = "CPU"
		}
		if strings.ContainsRune(p2, 0xfffd) {
			p2 = "CPU"
		}
		return p1, p2
	}()

	s := []byte("---------------------------------------------------------------------------")
	content = content[:getLastIndex(content, s)]
	content = content[getLastIndex(content, s)+len(s):]

	m := func() map[string]string {
		var infoMap map[string]string
		infoMap = make(map[string]string)

		y := strings.Split(string(content), "\r\n")

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

	return Info{
		Filename: filepath.Base(filename),
		Mode:     m["GameMode"],
		Time:     m["Time"],
		Winner:   m["Winner"],
		Player1: Player{
			Name:      player1,
			Character: m["1P-side"],
		},
		Player2: Player{
			Name:      player2,
			Character: m["2P-side"],
		},
	}
}

func mapOsArgs(handler func(a string)) {
	for _, v := range os.Args[1:] {
		handler(v)
	}
}

func (info Info) ToString() string {
	var sb strings.Builder
	newLine := "\n"
	if runtime.GOOS == "windows" {
		newLine = "\r\n"
	}

	sb.WriteString("Filename: ")
	sb.WriteString(info.Filename)
	sb.WriteString(newLine)

	if info.IsBadFile {
		sb.WriteString("Bad file cannot be read")
		sb.WriteString(newLine)
		return sb.String()
	}

	sb.WriteString("Time: ")
	sb.WriteString(info.Time)
	sb.WriteString(newLine)

	sb.WriteString("Mode: ")
	sb.WriteString(info.Mode)
	//sb.WriteString(" | Winner: ")
	//sb.WriteString(info.Winner)
	sb.WriteString(newLine)

	sb.WriteString("1P: ")
	sb.WriteString(info.Player1.Character)
	sb.WriteString(" ")
	sb.WriteString(info.Player1.Name)
	if info.Winner == "1P" {
		sb.WriteString(" [Winner]")
	}
	sb.WriteString(newLine)

	sb.WriteString("2P: ")
	sb.WriteString(info.Player2.Character)
	sb.WriteString(" ")
	sb.WriteString(info.Player2.Name)
	if info.Winner == "2P" {
		sb.WriteString(" [Winner]")
	}
	sb.WriteString(newLine)
	sb.WriteString(newLine)

	return sb.String()
}

func (info Info) IsEmpty() bool {
	return info == Info{}
}

func (infos Infos) SaveTXT(path string) error {
	var txtSB strings.Builder
	for _, v := range infos {
		txtSB.WriteString(v.ToString())
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	err = file.Truncate(0)
	_, err = file.Seek(0, 0)

	_, err = file.WriteString(txtSB.String())
	return err
}

func handleOsArg(filename string) {
	if isRepFile(filename) {
		if info := getReplayInfo(filename); !info.IsEmpty() {
			fmt.Print(info.ToString())
			_ = info
		} else {
			fmt.Print(Info{
				Filename:  filepath.Base(filename),
				IsBadFile: true,
			}.ToString())
		}
	}
	if isDir(filename) {
		infos, err := readRepDir(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		infos.SaveTXT(filename + "/" + "list.txt")
		fmt.Println("Result of directory", filename, "saved as list.txt")
		fmt.Println()
	}
}

func main() {
	fmt.Println("Replay Info:")
	fmt.Println("-------------------------")
	fmt.Println()

	mapOsArgs(handleOsArg)

	fmt.Println("-------------------------")

	func() {
		fmt.Println("Press Ctrl-C to exit...")
		for {
			time.Sleep(time.Second)
		}
	}()
}

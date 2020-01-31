package main

import (
	"fmt"
	"github.com/DanieleDaccurso/goxdo"
	"github.com/MarinX/keylogger"
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	//	"strconv"
	"encoding/json"
	"flag"
	//	"golang.org/x/text/unicode/rangetable"
	"os"
	"runtime"
	"strings"
	"time"
	//	"unicode"
)

type Expansion struct {
	Abbrev   string
	Expanded string
}

// Maybe replaceable by https://godoc.org/golang.org/x/mobile/event/key ?
var keys = map[rune]string{
	'a':  "a",
	'b':  "b",
	'c':  "c",
	'd':  "d",
	'e':  "e",
	'f':  "f",
	'g':  "g",
	'h':  "h",
	'i':  "i",
	'j':  "j",
	'k':  "k",
	'l':  "l",
	'm':  "m",
	'n':  "n",
	'o':  "o",
	'p':  "p",
	'q':  "q",
	'r':  "r",
	's':  "s",
	't':  "t",
	'u':  "u",
	'v':  "v",
	'w':  "w",
	'x':  "x",
	'y':  "y",
	'z':  "z",
	'A':  "A",
	'B':  "B",
	'C':  "C",
	'D':  "D",
	'E':  "E",
	'F':  "F",
	'G':  "G",
	'H':  "H",
	'I':  "I",
	'J':  "J",
	'K':  "K",
	'L':  "L",
	'M':  "M",
	'N':  "N",
	'O':  "O",
	'P':  "P",
	'Q':  "Q",
	'R':  "R",
	'S':  "S",
	'T':  "T",
	'U':  "U",
	'V':  "V",
	'W':  "W",
	'X':  "X",
	'Y':  "Y",
	'Z':  "Z",
	'1':  "1",
	'2':  "2",
	'3':  "3",
	'4':  "4",
	'5':  "5",
	'6':  "6",
	'7':  "7",
	'8':  "8",
	'9':  "9",
	'0':  "0",
	'-':  "minus",
	'_':  "underscore",
	'=':  "equal",
	'[':  "bracketleft",
	']':  "bracketright",
	'\\': "backslash",
	';':  "semicolon",
	',':  "comma",
	'.':  "period",
	'/':  "slash",
	' ':  "space",
	'\n': "Return",
	'!':  "exclam",
	'@':  "at",
	'#':  "numbersign",
	'$':  "dollar",
	'%':  "percent",
	'^':  "asciicircum",
	'&':  "ampersand",
	'*':  "asterisk",
	'(':  "parenleft",
	')':  "parenright",
	'<':  "less",
	'>':  "greater",
	'?':  "question",
	'"':  "quotedbl",
	'\'': "quoteright",
	'{':  "braceleft",
	'}':  "braceright",
	'~':  "asciitilde",
	'|':  "bar",
}

var verbose bool

func main() {

	var config string
	var device string

	flag.StringVar(&config, "c", fmt.Sprintf("%s/.macros", os.Getenv("HOME")), "Config file to use for macros.")
	flag.StringVar(&device, "d", "", "Input device to listen to. For input devices use: ls -al /dev/input/by-id")
	flag.BoolVar(&verbose, "v", false, "Enable verbose mode.")
	flag.Parse()

	if verbose {
		logrus.Println("Initializing...")
		logrus.Println("Using config file: ", config)
	}

	// TODO multi-keyboard
	kb, err := keybd_event.NewKeyBonding()
	check(err)

	// For linux, it is very important wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	if device == "" {
		device = keylogger.FindKeyboardDevice()
		if len(device) < 1 {
			logrus.Error("No keyboard found...you will need to provide manual input path")
			return
		}
	}

	if verbose {
		logrus.Println("Using keyboard at", device)
	}

	// init keylogger with keyboard
	k, err := keylogger.New(device)
	check(err)
	defer k.Close()

	events := k.Read()

	// arbitrarily set limit to 255
	var expansions []Expansion = make([]Expansion, 0, 255)
	file, err := os.OpenFile(config, os.O_CREATE, 0666)
	check(err)
	defer file.Close()

	if verbose {
		logrus.Println("Reading in config file ...")
	}
	var bytes []byte = make([]byte, 1000, 5000)
	read, err := file.Read(bytes)
	check(err)
	err = json.Unmarshal(bytes[:read], &expansions)
	check(err)

	if verbose {
		logrus.Println("Loaded macros: ", expansions)
	}
	var pressed = make([]string, 0, 50)

	logrus.Println("fox expander ready!")
	for e := range events {

		switch e.Type {
		// EvKey is used to describe state changes of keyboards, buttons, or other key-like devices.
		// check the input_event.go for more events
		case keylogger.EvKey:

			// if the state of key is pressed
			if e.KeyPress() {
				if len(pressed) == cap(pressed) {
					pressed = reset(pressed)
				}
				pressed = append(pressed, e.KeyString())
				match, exp := checkExpand(pressed, expansions)
				if match {
					expand(exp, kb)
					pressed = make([]string, 0, 50)
				}
				if verbose {
					logrus.Println("[event] press key ", e.KeyString())
					logrus.Println("Key Sequence ", pressed)
				}

			}
			break
		}
	}
}
func checkExpand(pressed []string, expansions []Expansion) (bool, Expansion) {
	// All abbreviations are checked case insensitive
	joined := strings.Join(pressed, "")
	// If we get a lot of these I will break out into seperate function
	joined = strings.ToLower(strings.Replace(joined, "SPACE", " ", -1))

	if verbose {
		logrus.Printf("Checking for matches in: '%s'", joined)
	}
	for _, exp := range expansions {
		if strings.Contains(joined, strings.ToLower(exp.Abbrev)) {
			if verbose {
				logrus.Printf("Match found: '%s'!\n Expanding...\n", exp.Abbrev)
			}
			return true, exp
		}
	}
	// Not sure how to best handle the return of this function
	return false, Expansion{"", ""}
}

func expand(exp Expansion, kb keybd_event.KeyBonding) {

	if verbose {
		logrus.Printf("Expanding %s", exp.Expanded)
	}
	// Insert a backspace for each rune in the Abbreviation
	for i := 0; i < len(exp.Abbrev); i++ {
		kb.SetKeys(keybd_event.VK_BACKSPACE)
		err := kb.Launching()
		time.Sleep(8 * time.Millisecond)
		check(err)
	}
	kb.Clear()

	// Sleeps are required as theres an apparent race condition.
	// Keys launched first are not guaranteed to be typed first
	xdo_slice := make([]string, 0, 10000)
	for _, char := range exp.Expanded {
		xdo_slice = append(xdo_slice, keys[char])
	}
	xdo_string := strings.Join(xdo_slice, " ")

	if verbose {
		logrus.Printf("xdotool string built:\n%s", xdo_string)
	}
	xdo := goxdo.NewXdo()
	xdo.EnterTextWindow(xdo.GetWindowAtMouse(), exp.Expanded, 5)
	time.Sleep(12 * time.Millisecond)

	kb.Clear()

}
func reset(current []string) []string {

	// 50 is arbitrary here and may need to increase
	if verbose {
		logrus.Println("Keys reset.")
	}
	temp := make([]string, 0, 50)
	temp = append(temp, current[40:]...)
	return temp
}
func check(e error) {
	if e != nil {
		logrus.Fatal(e)
	}
}

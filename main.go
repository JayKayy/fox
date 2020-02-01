package main

import (
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	//	"strconv"
	"encoding/json"
	"flag"
	"golang.org/x/text/unicode/rangetable"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode"
)

type Expansion struct {
	Abbrev   string
	Expanded string
}

// Maybe replaceable by https://godoc.org/golang.org/x/mobile/event/key ?
var keys = map[rune]int{
	'a':  keybd_event.VK_A,
	'b':  keybd_event.VK_B,
	'c':  keybd_event.VK_C,
	'd':  keybd_event.VK_D,
	'e':  keybd_event.VK_E,
	'f':  keybd_event.VK_F,
	'g':  keybd_event.VK_G,
	'h':  keybd_event.VK_H,
	'i':  keybd_event.VK_I,
	'j':  keybd_event.VK_J,
	'k':  keybd_event.VK_K,
	'l':  keybd_event.VK_L,
	'm':  keybd_event.VK_M,
	'n':  keybd_event.VK_N,
	'o':  keybd_event.VK_O,
	'p':  keybd_event.VK_P,
	'q':  keybd_event.VK_Q,
	'r':  keybd_event.VK_R,
	's':  keybd_event.VK_S,
	't':  keybd_event.VK_T,
	'u':  keybd_event.VK_U,
	'v':  keybd_event.VK_V,
	'w':  keybd_event.VK_W,
	'x':  keybd_event.VK_X,
	'y':  keybd_event.VK_Y,
	'z':  keybd_event.VK_Z,
	'A':  keybd_event.VK_A,
	'B':  keybd_event.VK_B,
	'C':  keybd_event.VK_C,
	'D':  keybd_event.VK_D,
	'E':  keybd_event.VK_E,
	'F':  keybd_event.VK_F,
	'G':  keybd_event.VK_G,
	'H':  keybd_event.VK_H,
	'I':  keybd_event.VK_I,
	'J':  keybd_event.VK_J,
	'K':  keybd_event.VK_K,
	'L':  keybd_event.VK_L,
	'M':  keybd_event.VK_M,
	'N':  keybd_event.VK_N,
	'O':  keybd_event.VK_O,
	'P':  keybd_event.VK_P,
	'Q':  keybd_event.VK_Q,
	'R':  keybd_event.VK_R,
	'S':  keybd_event.VK_S,
	'T':  keybd_event.VK_T,
	'U':  keybd_event.VK_U,
	'V':  keybd_event.VK_V,
	'W':  keybd_event.VK_W,
	'X':  keybd_event.VK_X,
	'Y':  keybd_event.VK_Y,
	'Z':  keybd_event.VK_Z,
	'1':  keybd_event.VK_1,
	'2':  keybd_event.VK_2,
	'3':  keybd_event.VK_3,
	'4':  keybd_event.VK_4,
	'5':  keybd_event.VK_5,
	'6':  keybd_event.VK_6,
	'7':  keybd_event.VK_7,
	'8':  keybd_event.VK_8,
	'9':  keybd_event.VK_9,
	'0':  keybd_event.VK_0,
	'-':  keybd_event.VK_SP2,
	'=':  keybd_event.VK_SP3,
	'[':  keybd_event.VK_SP4,
	']':  keybd_event.VK_SP5,
	'\\': keybd_event.VK_SP8,
	';':  keybd_event.VK_SP6,
	'\'': keybd_event.VK_SP7,
	',':  keybd_event.VK_SP9,
	'.':  keybd_event.VK_SP10,
	'/':  keybd_event.VK_SP11,
	' ':  keybd_event.VK_SPACE,
	'\n': keybd_event.VK_ENTER,
	'!':  keybd_event.VK_1,
	'@':  keybd_event.VK_2,
	'#':  keybd_event.VK_3,
	'$':  keybd_event.VK_4,
	'%':  keybd_event.VK_5,
	'^':  keybd_event.VK_6,
	'&':  keybd_event.VK_7,
	'*':  keybd_event.VK_8,
	'(':  keybd_event.VK_9,
	')':  keybd_event.VK_0,
}

var verbose bool
var delay int32 = 20

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

	// Insert a backspace for each rune in the Abbreviation
	for i := 0; i < len(exp.Abbrev); i++ {
		kb.SetKeys(keybd_event.VK_BACKSPACE)
		err := kb.Launching()
		time.Sleep(time.Duration(rand.Int31n(delay)) * time.Millisecond)
		check(err)
	}
	kb.Clear()

	// Sleeps are required as theres an apparent race condition.
	// Keys launched first are not guaranteed to be typed first
	for _, char := range exp.Expanded {
		if unicode.IsLetter(char) {
			if verbose {
				logrus.Printf("Expanding char as letter '%c'...", char)
			}
			if unicode.IsUpper(char) {
				kb.HasSHIFT(true)
			}
		} else if unicode.IsPunct(char) {
			if verbose {
				logrus.Printf("Expanding char as punctuation '%c'...", char)
			}
			if unicode.In(char, rangetable.New('!', '@', '#', '$', '%', '^', '&', '*', '(', ')')) {
				if verbose {
					logrus.Printf("Activating shift for '%c'...", char)
				}
				kb.HasSHIFT(true)
			}
		} else {
			if verbose {
				logrus.Printf("Expanding char as non letter or punctuation '%c'...", char)
			}
		}
		kb.SetKeys(keys[char])
		err := kb.Launching()
		check(err)
		time.Sleep(time.Duration(rand.Int31n(delay)) * time.Millisecond)
		kb.Clear()
	}
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

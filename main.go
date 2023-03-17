package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/unicode/rangetable"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode"
)

// 3/16/2023
// This program can work but at this point relies too heavily on the keylogger's single key press process.
// I am unable to get it to reliably enter characters without introducing significant delays in
// outputting the results. Things to revisit:
// String implementation currently we output 1 char at a time for a reason I can't remamber.
// If we could say simply send "hello" to the keyboard that would likely be better.

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
var delay = 40

func main() {

	var config string
	var device string

	flag.StringVar(&config, "c", fmt.Sprintf("%s/.macros", os.Getenv("HOME")), "Config file to use for macros.")
	flag.StringVar(&device, "d", "", "Input device to listen to. For input devices use: ls -al /dev/input/by-id")
	flag.BoolVar(&verbose, "v", false, "Enable verbose mode.")
	flag.Parse()

	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Info("Initializing...")
	logrus.Debug("Using config file: ", config)

	// TODO multi-keyboard
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		logrus.Fatal("Setting up new KeyBonding", err)
	}
	// For linux, it is very important wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	if device == "" {
		device = keylogger.FindKeyboardDevice()
		if len(device) < 1 {
			logrus.Error("No keyboard found...you will need to provide manual input path. Ex: fox -d /dev/input/event[NUM]")
			return
		}
	}

	logrus.Debug("Using keyboard at", device)

	// init keylogger with keyboard
	k, err := keylogger.New(device)
	if err != nil {
		logrus.Fatal("initializing keylogger on device", err)
	}
	defer k.Close()

	events := k.Read()

	// arbitrarily set limit to 255
	var expansions = make([]Expansion, 0, 255)
	file, err := os.OpenFile(config, os.O_CREATE, 0666)
	if err != nil {
		logrus.Fatal("opening config file", err)
	}
	defer file.Close()

	logrus.Debug("Reading in config file ...")

	var bytes = make([]byte, 1000, 5000)
	read, err := file.Read(bytes)
	if err != nil {
		logrus.Fatalf("reading from config file %s: %v", config, err)
	}

	err = json.Unmarshal(bytes[:read], &expansions)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Debug("Loaded macros: ", expansions)

	var pressed = make([]string, 0, 50)

	logrus.Info("fox expander ready!")
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
				logrus.Debug("[event] press key ", e.KeyString())
				match, exp := checkExpand(pressed, expansions)
				if match {
					expand(exp, kb)
					pressed = make([]string, 0, 50)
				}
				//	logrus.Debug("Key Sequence ", pressed)
			}
			break
		}
	}
}
func checkExpand(pressed []string, expansions []Expansion) (bool, Expansion) {
	// All abbreviations are checked case-insensitive
	joined := strings.Join(pressed, "")
	// If we get a lot of these I will break out into seperate function
	joined = strings.ToLower(strings.Replace(joined, "SPACE", " ", -1))

	logrus.Debugf("Checking for matches in: '%s'", joined)

	for _, exp := range expansions {
		if strings.Contains(joined, strings.ToLower(exp.Abbrev)) {
			logrus.Debugf("Match found: '%s'!", exp.Abbrev)
			return true, exp
		}
	}
	return false, Expansion{"", ""}
}

func expand(exp Expansion, kb keybd_event.KeyBonding) {

	// Insert a backspace for each rune in the Abbreviation
	for i := 0; i < len(exp.Abbrev); i++ {
		logrus.Debug("Backspace pressed")
		kb.SetKeys(keybd_event.VK_BACKSPACE)
		err := kb.Launching()
		time.Sleep(time.Millisecond * time.Duration(delay))
		if err != nil {
			logrus.Fatal("expanding expression", err)
		}
	}
	kb.Clear()

	// Sleeps are required as there's a race condition.
	// Keys launched first are not guaranteed to be typed first
	for _, char := range exp.Expanded {
		if unicode.IsLetter(char) {
			logrus.Debugf("Expanding char as letter '%c'", char)
			if unicode.IsUpper(char) {
				kb.HasSHIFT(true)
			}
		} else if unicode.IsPunct(char) {
			logrus.Debugf("Expanding char as punctuation '%c'", char)
			if unicode.In(char, rangetable.New('!', '@', '#', '$', '%', '^', '&', '*', '(', ')')) {
				logrus.Debugf("Activating shift for '%c'...", char)
				kb.HasSHIFT(true)
			}
		} else {
			logrus.Debugf("Expanding char as non letter or punctuation '%c'", char)
		}
		kb.SetKeys(keys[char])
		err := kb.Launching()
		time.Sleep(time.Millisecond * time.Duration(delay))
		if err != nil {
			logrus.Fatal("launching macro", err)
		}
		kb.Clear()
	}
}
func reset(current []string) []string {

	// 50 is arbitrary here and may need to increase
	logrus.Debug("Keys reset.")

	temp := make([]string, 0, 50)
	temp = append(temp, current[40:]...)
	return temp
}

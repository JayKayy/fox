package main

import (
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	//	"strconv"
	"golang.org/x/text/unicode/rangetable"
	"runtime"
	"strings"
	"time"
	"unicode"
)

type Expansion struct {
	abbrev   string
	expanded string
}

// Maybe replaceable by https://godoc.org/golang.org/x/mobile/event/key ?
var keys = map[rune]int{
	'a': keybd_event.VK_A,
	'b': keybd_event.VK_B,
	'c': keybd_event.VK_C,
	'd': keybd_event.VK_D,
	'e': keybd_event.VK_E,
	'f': keybd_event.VK_F,
	'g': keybd_event.VK_G,
	'h': keybd_event.VK_H,
	'i': keybd_event.VK_I,
	'j': keybd_event.VK_J,
	'k': keybd_event.VK_K,
	'l': keybd_event.VK_L,
	'm': keybd_event.VK_M,
	'n': keybd_event.VK_N,
	'o': keybd_event.VK_O,
	'p': keybd_event.VK_P,
	'q': keybd_event.VK_Q,
	'r': keybd_event.VK_R,
	's': keybd_event.VK_S,
	't': keybd_event.VK_T,
	'u': keybd_event.VK_U,
	'v': keybd_event.VK_V,
	'w': keybd_event.VK_W,
	'x': keybd_event.VK_X,
	'y': keybd_event.VK_Y,
	'z': keybd_event.VK_Z,
	'A': keybd_event.VK_A,
	'B': keybd_event.VK_B,
	'C': keybd_event.VK_C,
	'D': keybd_event.VK_D,
	'E': keybd_event.VK_E,
	'F': keybd_event.VK_F,
	'G': keybd_event.VK_G,
	'H': keybd_event.VK_H,
	'I': keybd_event.VK_I,
	'J': keybd_event.VK_J,
	'K': keybd_event.VK_K,
	'L': keybd_event.VK_L,
	'M': keybd_event.VK_M,
	'N': keybd_event.VK_N,
	'O': keybd_event.VK_O,
	'P': keybd_event.VK_P,
	'Q': keybd_event.VK_Q,
	'R': keybd_event.VK_R,
	'S': keybd_event.VK_S,
	'T': keybd_event.VK_T,
	'U': keybd_event.VK_U,
	'V': keybd_event.VK_V,
	'W': keybd_event.VK_W,
	'X': keybd_event.VK_X,
	'Y': keybd_event.VK_Y,
	'Z': keybd_event.VK_Z,
	'1': keybd_event.VK_1,
	'2': keybd_event.VK_2,
	'3': keybd_event.VK_3,
	'4': keybd_event.VK_4,
	'5': keybd_event.VK_5,
	'6': keybd_event.VK_6,
	'7': keybd_event.VK_7,
	'8': keybd_event.VK_8,
	'9': keybd_event.VK_9,
	'0': keybd_event.VK_0,
	'-': keybd_event.VK_SP2,
	'=': keybd_event.VK_SP3,
	'[': keybd_event.VK_SP4,
	']': keybd_event.VK_SP5,
	//	'\\' keybd_event.VK_SP8,
	';': keybd_event.VK_SP6,
	//	'\'':  keybd_event.VK_SP7,
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

func main() {

	// Prep for send keys

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}
	// For linux, it is very important wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	// find keyboard device, does not require a root permission
	keyboard := keylogger.FindKeyboardDevice()
	// check if we found a path to keyboard
	if len(keyboard) <= 0 {
		logrus.Error("No keyboard found...you will need to provide manual input path")
		return
	}

	logrus.Println("Found a keyboard at", keyboard)
	// init keylogger with keyboard
	// trying for logitech keyboard
	// Otherwise it only picks up laptop keyboard
	k, err := keylogger.New(keyboard)
	//k, err := keylogger.New("/dev/input/event19")
	if err != nil {
		logrus.Error(err)
		return
	}
	defer k.Close()

	events := k.Read()

	// range of events
	// specify abbreviations for expansion
	// TODO load them in from filesystem
	// arbitrarily set limit to 255
	var expansions []Expansion = make([]Expansion, 0, 255)

	// Set static expansions
	expansions = append(expansions, Expansion{";;,", "Hi there!\n\nThanks for contacting DigitalOcean!"})
	expansions = append(expansions, Expansion{";;.", "Hi there!\n\nThank you for your response!"})
	expansions = append(expansions, Expansion{";;/", "Let me know if you have any further questions.\n\nRegards,\n\nJohn Kwiatkoski\nSenior Developer Support Engineer - Kubernetes"})

	var pressed = make([]string, 0, 50)
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
				logrus.Println("[event] press key ", e.KeyString())
				pressed = append(pressed, e.KeyString())
				match, exp := checkExpand(pressed, expansions)
				if match {
					expand(exp, kb)
					pressed = make([]string, 0, 50)
				}
				fmt.Println(pressed)

			}

			// if the state of key is released
			// Dont think ill need releases
			//	if e.KeyRelease() {
			//		logrus.Println("[event] release key ", e.KeyString())
			//	}

			break
		}
	}
}
func checkExpand(pressed []string, expansions []Expansion) (bool, Expansion) {

	// compact pressed to a single string
	// check whether any abbrieviations (abbrevs)
	// are in the keys that were pressed.
	// If so return the abbrev
	joined := strings.Join(pressed, "")
	for _, exp := range expansions {
		if strings.Contains(joined, exp.abbrev) {
			fmt.Printf("Match found: %s!\n Expanding...\n", exp.abbrev)
			return true, exp
		}
	}
	// Not sure how to best handle the return of this function
	return false, Expansion{"", ""}
}

func expand(exp Expansion, kb keybd_event.KeyBonding) {

	// Cant do all in one because of caps and shifts
	// var keyed_msg = make([]int, 0, 4000)

	// Insert a backspace for each rune in the abbreviation
	for i := 0; i < len(exp.abbrev); i++ {
		kb.SetKeys(keybd_event.VK_BACKSPACE)
		err := kb.Launching()
		time.Sleep(14 * time.Millisecond)
		if err != nil {
			panic(err)
		}
	}
	kb.Clear()

	for _, char := range exp.expanded {
		if unicode.IsLetter(char) {
			if unicode.IsUpper(char) {
				kb.HasSHIFT(true)
			}
		}
		if unicode.IsPunct(char) {
			if unicode.In(char, rangetable.New('!', '@', '#', '$', '%', '^', '&', '*', '(', ')')) {
				kb.HasSHIFT(true)
			}
		}
		kb.SetKeys(keys[char])
		err := kb.Launching()
		if err != nil {
			panic(err)
		}
		time.Sleep(8 * time.Millisecond)
		kb.Clear()

	}
	//launch
	//Ouput : AB
}
func reset(current []string) []string {

	// 40 is arbitrary here and may need to move if
	// the slice is ever made bigger
	fmt.Println("keys reset.")
	temp := make([]string, 0, 50)
	temp = append(temp, current[40:]...)
	return temp
}

package main

import (
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/sirupsen/logrus"
	"strings"
)

type Expansion struct {
	abbrev   string
	expanded string
}

func main() {

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
	k, err := keylogger.New("/dev/input/event19")
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
	expansions = append(expansions, Expansion{";;,", "Hi there!\n Thanks for contacting DigitalOcean!"})

	var pressed = make([]string, 0, 50)
	for e := range events {

		switch e.Type {
		// EvKey is used to describe state changes of keyboards, buttons, or other key-like devices.
		// check the input_event.go for more events
		case keylogger.EvKey:

			// if the state of key is pressed
			if e.KeyPress() {
				if len(pressed) == cap(pressed) {
					// TODO I want the new array to begin with 40,41,42,43...
					// zero out 0-39

					for i := 0; i < 40; i++ {
						pressed[i] = ""
					}
					fmt.Println("zeroed out")
					temp := make([]string, 0, 50)
					fmt.Printf("before copy pressed[40:] is: %s", pressed[40:])
					n := copy(temp, pressed[40:])
					fmt.Printf("%d copied. temp is now: %s", n, temp)
					pressed = temp
					fmt.Println(pressed)
				}
				logrus.Println("[event] press key ", e.KeyString())
				pressed = append(pressed, e.KeyString())
				pressed = checkExpand(pressed, expansions)
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
func checkExpand(pressed []string, expansions []Expansion) []string {

	// compact pressed to a single string
	// check whether any abbrieviations (abbrevs)
	// are in the keys that were pressed.
	// If so return the abbrev
	joined := strings.Join(pressed, "")
	for _, exp := range expansions {
		if strings.Contains(joined, exp.abbrev) {
			fmt.Printf("Match found in %s!", exp.abbrev)
		}
	}
	return pressed
}

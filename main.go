package main

import (
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	//	"runtime"
	"strings"
	//	"time"
)

type Expansion struct {
	abbrev   string
	expanded string
}

func main() {

	// Prep for send keys

	kb, err := keybd_event.NewKeyBonding()

	if err != nil {
		panic(err)
	}

	// Ignoring this warning from micmonay, YOLO!
	// We wont be sending keys immediately
	// For linux, it is very important wait 2 seconds
	//if runtime.GOOS == "linux" {
	//	time.Sleep(2 * time.Second)
	//}

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
	expansions = append(expansions, Expansion{";;.", "Hi there!\n Thank you for your response!"})
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
			fmt.Printf("Match found in %s! Expanding...", exp.abbrev)
			return true, exp
		}
	}
	// Not sure how to best handle the return of this function
	return false, Expansion{"", ""}
}

func expand(exp Expansion, kb keybd_event.KeyBonding) {

	//Built expansion here
	kb.SetKeys(keybd_event.VK_A, keybd_event.VK_B)
	//launch
	err := kb.Launching()
	if err != nil {
		panic(err)
	}
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

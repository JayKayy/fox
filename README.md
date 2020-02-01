# Fox - Text Expander

### Overview

`fox` is a lightweight CLI text expander for linux meant to help reduce 
repetitive typing. It can be used for anything from templating to email 
greetings and signatures.

### Usage

```
Usage of ./fox:
  -c string
    	Config file to use for macros. (default "$HOME/.macros")
  -d string
    	Input device to listen to. For input devices use: ls -al /dev/input/by-id
  -v	Enable verbose mode.
```

Since fox reads keyboard presses, it requires escalated privileges to run or
opening up privileges on `/dev/uinput`.

### Macros

A sample macro file (macros) can be found in this repo. It is a json formatted 
text file that contains a list of the abbreviations and what they should
expand to. You can add up to 255 macros to the list. The default macro file 
used will be $HOME/.macros. Otherwise you can specify the path to your macro 
file with the `-c` flag. 

### Devices

Only one device (keyboard) can be used at a time. A default device will be 
chosen if one is not specified. To list all your devices, you can use the 
command `ls -al /dev/input/by-id`. Then providing that path to the `-d` flag
will set `fox` to listen for keystrokes on that specific device.

### Examples

Use default keyboard and `/root/.macros` file

```bash
sudo ./fox
```

Use default keyboard and `/home/userA/.macros`

```bash
sudo ./fox -c /home/userA/.macros
```

Use USB keyboard and `/home/userA/.macros`

```bash
sudo ./fox -c /home/userA/.macros -d /dev/input/event17
```

Use default keyboard and `/home/userA/.macros` and print verbose output.

```bash
sudo ./fox -c /home/userA/.macros -d /dev/input/event17 -v 
```


# Fox - Text Expander

### Overview

`fox` is a lightweight CLI text expander for linux meant
to help reduce repetitive typing. It can be used for anything from
templating to email greetings and signatures.

### Usage

```
Fox - Text Expander 0.1.0
Expands text abbreviations

USAGE:
    fox [OPTIONS]

FLAGS:
    -h, --help       Prints help information
    -V, --version    Prints version information

OPTIONS:
    -c, --config <config>    json file containing list of macros
    -d, --device <device>    Path to device file. Normally in /dev/input
```

Since fox reads keyboard presses, it requires escalated privileges to run or
opening up privileges on `/dev/input`.

### Macros

A sample macro file (macros) can be found in this repo. It is a json formatted 
text file that contains a list of the abbreviations and what they should
expand to once typed. The default macro file 
used will be `$HOME/.config/macros`. Otherwise you can specify the path to your
macro file with the `-c` flag. 

### Devices

Only one device (keyboard) can be used at a time. To list all your devices,
you can use the command `ls -al /dev/input/by-id`. Then providing that path
to the `-d` flag will set fox to listen to the provided device for keystrokes.

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

Use default keyboard and `/home/userA/.macros`.

```bash
sudo ./fox -c /home/userA/.macros -d /dev/input/event17
```


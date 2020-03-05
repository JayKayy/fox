extern crate serde_json;
extern crate serde;
#[macro_use]
extern crate serde_derive;
extern crate clap;

mod input;
use input::{is_key_event, is_key_press, is_key_release, is_shift, get_key_text, InputEvent};

//use getopts::Options;
use std::fs::File;
use std::io::Read;
use std::mem;
use enigo::*;
use clap::{Arg, App};

#[derive(Serialize, Deserialize, Debug)]
struct Expansion {
    abbrev: String,
    expanded: String
}

#[derive(Serialize, Deserialize, Debug)]
struct ExpansionSet {
    expansions: Vec<Expansion>
}


fn main() -> std::io::Result<()> {

    let matches = App::new("Fox - Text Expander")
        .version("0.1.0")
        .about("Expands text abbreviations")
        .arg(Arg::with_name("config")
            .short("c")
            .long("config")
            .takes_value(true)
            .help("json file containing list of macros"))
        .arg(Arg::with_name("device")
            .short("d")
            .long("device")
            .takes_value(true)
            .required(true)
            .help("Path to device file. Normally in /dev/input"))
        .get_matches();

    // Specify event to listen to, bind multiple?
    let default_device = "";
    let default_config = "$HOME/.config/macros";

    let config_path = matches.value_of("config").unwrap_or(default_config);
    let device_path = matches.value_of("device").unwrap_or(default_device);
    let mut device = setup_device(device_path);
    let expset = load_expansions(config_path);

    println!("Initialization Complete!");

    // Loop through keyboard events
    let mut buffer = [0;24];
    let mut pressed = Vec::new();
    let mut shift_pressed = 0;
    loop {
        let num_bytes = device.read(&mut buffer).unwrap_or_else(|e| panic!("{}", e));
        if num_bytes != mem::size_of::<InputEvent>() {
            panic!("Error while reading from device file");
        }
        let event: InputEvent = unsafe { mem::transmute(buffer) };
        if is_key_event(event.type_) {
            if is_key_press(event.value) {
                if is_shift(event.code) {
                    shift_pressed += 1;
                }

                let text = get_key_text(event.code, shift_pressed);
                pressed.push(text);
                if check_expand(&expset, &pressed) {
                    pressed.clear();
                }

            } else if is_key_release(event.value) {
                if is_shift(event.code) {
                    shift_pressed -= 1;
                }
            }
        }
    }
}

fn setup_device(device: &str) -> std::fs::File {
    // Setup Device to listen to
    let file = File::open(device);
    if file.is_err(){
        panic!("Error opening device: {}!", device);
    }
    return file.unwrap();
}

fn load_expansions(default_config: &str) -> ExpansionSet {
    // Read in ExpansionSet
    let config = File::open(default_config);
    if config.is_err(){
        panic!("Error opening macros file {}", default_config);
    }
    let res = serde_json::from_reader(config.unwrap());
    if res.is_err() {
        panic!("Error parsing JSON in: {}", default_config);
    }
    return res.unwrap();
}


// Simulate keypresses
fn expand(exp: &Expansion) -> bool {
    let mut enigo = Enigo::new();
    for _i in 0..exp.abbrev.len() {
        enigo.key_down(Key::Backspace);
        enigo.key_up(Key::Backspace);
    }
    // DSL seems broken need to split on '\n' and print each line
    if exp.expanded.contains("\n") {
        let mut i = 0;
        for line in exp.expanded.split("\n") {
            enigo.key_sequence(&(line));
            // Don't add a new line at the end of the last line.
            if i < exp.expanded.split("\n").collect::<Vec<&str>>().len()-1 {
                enigo.key_down(Key::Return);
                enigo.key_up(Key::Return);
            }
            i = i+1;
        }
    }else {
        enigo.key_sequence(&exp.expanded);
    }
    return true
}

// Check for abbreviation matches
fn check_expand(set: &ExpansionSet, pressed: &Vec<&str>) -> bool {
    let string = pressed.join("");
    for exp in set.expansions.iter() {
        if string.contains(&exp.abbrev){
            return expand(&exp)
        }
    }
    return false
}

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
use std::{thread, time};
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
        .arg(Arg::with_name("verbose")
            .short("v")
            .long("verbose")
            .required(false)
            .help("Enables verbose output"))
        .get_matches();

    // Specify event to listen to, bind multiple?
    let default_device = "";
    let default_config = "$HOME/.config/macros";

    let config_path = matches.value_of("config").unwrap_or(default_config);
    let device_path = matches.value_of("device").unwrap_or(default_device);
    let verbose = matches.is_present("verbose");
    if verbose {
        println!("config: {}", config_path);
        println!("device_path: {}", device_path);
    }
    let mut device = setup_device(device_path, verbose);
    let expset = load_expansions(config_path, verbose);

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

            } else if is_key_release(event.value) {
                if is_shift(event.code) {
                    shift_pressed -= 1;
                }
                let text = get_key_text(event.code, shift_pressed);
                if verbose {
                    println!("Key pressed: {:?}", text);
                }
                pressed.push(text);
                if verbose {
                    println!("Buffer: {:?}", pressed);
                }
                           
                if check_expand(&expset, &pressed, verbose) {
                    if verbose {
                        println!("Buffer about to clear! Buffer: {:?}", pressed);
                    }
                    pressed.clear();
                    if verbose {
                        println!("Buffer cleared! Buffer: {:?}", pressed);
                    }
                }
            }
        }
    }
}

fn setup_device(device: &str, verbose: bool) -> std::fs::File {
    // Setup Device to listen to
    let file = File::open(device);
    if verbose {
        println!("Opening device.. {}", device);
    }
    if file.is_err(){
        panic!("Error opening device: {}!", device);
    }
    if verbose {
        println!("Device access complete.");
    }
    return file.unwrap();
}

fn load_expansions(default_config: &str, verbose: bool) -> ExpansionSet {
    // Read in ExpansionSet
    let config = File::open(default_config);
    if verbose {
        println!("Opened {}", default_config);
    }
    if config.is_err(){
        panic!("Error opening macros file {}", default_config);
    }
    let res = serde_json::from_reader(config.unwrap());
    if res.is_err() {
        panic!("Error parsing JSON in: {}", default_config);
    }
    if verbose {
        println!("Configuration parsing complete");
    }
    return res.unwrap();
}


// Simulate keypresses
fn expand(exp: &Expansion, verbose: bool) -> bool {
    let mut enigo = Enigo::new();
    if verbose {
        println!("Enigo created.");
    }
    for i in 0..exp.abbrev.len() {
        if verbose {
            println!("Pressing Backspace. {}", i);
        }
        // Combat race condition leaving behind parts of macro
        let ten_millis = time::Duration::from_millis(30);
        thread::sleep(ten_millis);
        enigo.key_click(Key::Backspace);
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
fn check_expand(set: &ExpansionSet, pressed: &Vec<&str>, verbose: bool) -> bool {
    let string = pressed.join("");
    for exp in set.expansions.iter() {
        if string.contains(&exp.abbrev){
            return expand(&exp, verbose);
        }
    }
    return false
}

extern crate serde_json;
extern crate serde;
#[macro_use]
extern crate serde_derive;

mod input;
use input::{is_key_event, is_key_press, is_key_release, is_shift, get_key_text, InputEvent};

//use getopts::Options;
use std::fs::File;
use std::io::Read;
use std::mem;
use enigo::*;

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

    // Specify event to listen to, bind multiple?
    let default_device = "/dev/input/event3";
    let default_config = "/home/jkwiatko/.config/macros";
    //TODO flags for specifying config file and device

    let mut device = setup_device(default_device);
    let expset = load_expansions(default_config);

    println!("Initialization Complete!");
    println!("Expansions: {}", expset.expansions[0].expanded);

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

fn setup_device(default_device: &str) -> std::fs::File {
    // Setup Device to listen to
    let file = File::open(default_device);
    if file.is_err(){
        panic!("Error opening default device!");
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
    enigo.key_sequence(&(exp.expanded));
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

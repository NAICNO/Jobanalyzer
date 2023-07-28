// Read system configuration data from a json file into a hashmap with the host name as key.
//
// The file format is an array [...] of objects { ... }, each with the following named fields and
// value types:
//
//   hostname - string, the fully qualified and unique host name of the node
//   description - string, optional, arbitrary text describing the system
//   cpu_cores - integer, the number of hyperthreads
//   mem_gb - integer, the amount of main memory in gigabytes
//   gpu_cards - integer, the number of gpu cards on the node
//   gpu_mem_gb - integer, the amount of gpu memory in gigabytes across all cards
//
// See ../ml-systems.json for an example.

use anyhow::{bail,Result};
use serde_json::Value;
use std::collections::HashMap;
use std::fs::File;
use std::io::BufReader;
use std::path;

// See above comment block for field documentation.

#[derive(Debug,Default)]
pub struct System {
    pub hostname: String,
    pub description: String,
    pub cpu_cores: usize,
    pub mem_gb: usize,
    pub gpu_cards: usize,
    pub gpu_mem_gb: usize
}

// Returns a map from host name to config info, or an error message.

// Since the input is human-generated, may vary a bit over time, and have optional fields, I've
// opted to use the generic JSON parser followed by explicit decoding of the fields, rather than a
// (derived) strongly-typed parser.

pub fn read_from_json(filename: &str) -> Result<HashMap<String, System>> {
    let file = File::open(path::Path::new(filename))?;
    let reader = BufReader::new(file);
    let v = serde_json::from_reader(reader)?;
    let mut m = HashMap::new();
    if let Value::Array(objs) = v {
        for obj in objs {
            if let Value::Object(fields) = obj {
                let mut sys : System = Default::default();
                if let Some(Value::String(hn)) = fields.get("hostname") {
                    sys.hostname = hn.clone();
                } else {
                    bail!("Field 'hostname' must be present and have a string value");
                }
                if let Some(d) = fields.get("description") {
                    if let Value::String(desc) = d {
                        sys.description = desc.clone();
                    } else {
                        bail!("Field 'description' must have a string value");
                    }
                }
                sys.cpu_cores = grab_usize(&fields, "cpu_cores")?;
                sys.mem_gb = grab_usize(&fields, "mem_gb")?;
                sys.gpu_cards = grab_usize(&fields, "gpu_cards")?;
                sys.gpu_mem_gb = grab_usize(&fields, "gpu_mem_gb")?;
                let key = sys.hostname.clone();
                m.insert(key, sys);
            } else {
                bail!("Expected an object value")
            }
        }
    } else {
        bail!("Expected an array value")
    }
    Ok(m)
}

fn grab_usize(fields: &serde_json::Map<String,Value>, name: &str) -> Result<usize> {
    if let Some(Value::Number(cores)) = fields.get(name) {
        if let Some(n) = cores.as_u64() {
            // TODO: Assert it fits in usize
            Ok(n as usize)
        } else {
            bail!("Field '{name}' must have unsigned integer value")
        }
    } else {
        bail!("Field '{name}' must be present and have an integer value")
    }
}

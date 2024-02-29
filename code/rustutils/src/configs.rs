/// Read system configuration data from a json file into a map with the host name as key.
///
/// The current file format is an array [...] of objects { ... }, each with the following named
/// fields and value types:
///
///   timestamp - string, an RFC3339 timestamp for when the data were obtained
///   hostname - string, the fully qualified and unique host name of the node
///   description - string, optional, arbitrary text describing the system
///   comment - string, optional, arbitrary text for non-public documentation use
///   cross_node_jobs - bool, optional, expressing that jobs on this node can be merged with
///                     jobs on other nodes in the same cluster where the flag is also set,
///                     because the job numbers come from the same cluster-wide source
///                     (typically slurm).  Also see the --batch option.
///   cpu_cores - integer, the number of hyperthreads
///   mem_gb - integer, the amount of main memory in gigabytes
///   gpu_cards - integer, the number of gpu cards on the node
///   gpumem_gb - integer, the amount of gpu memory in gigabytes across all cards
///   gpumem_pct - bool, optional, expressing a preference for the GPU memory reading
///
/// See ../../production/sonar-nodes/mlx.hpc.uio.no/mlx.hpc.uio.no-config.json for an example.
///
/// This is a partial implementation.  There will be another method on ClusterConfig to lookup by
/// host name and timestamp, and the data structure and file format will be more complicated to
/// support that.
use anyhow::{bail, Result};
use serde_json::Value;
use std::collections::HashMap;
use std::fs::File;
use std::io::BufReader;
use std::path;
use std::rc::Rc;

// See above comment block for field documentation.

#[derive(Debug, Default, Clone)]
pub struct System {
    pub timestamp: String,
    pub hostname: String,
    pub description: String,
    pub comment: String,
    pub cross_node_jobs: bool,
    pub cpu_cores: usize,
    pub mem_gb: usize,
    pub gpu_cards: usize,
    pub gpumem_gb: usize,
    pub gpumem_pct: bool,
}

pub struct ClusterConfig {
    nodes: HashMap<String, Rc<System>>,
}

impl ClusterConfig {
    pub fn lookup(&self, hostname: &str) -> Option<Rc<System>> {
        self.nodes.get(hostname).cloned()
    }

    pub fn cross_node_jobs(&self) -> bool {
        self.nodes.iter().any(|(_, sys)| sys.cross_node_jobs)
    }
}

/// Returns a map from host name to config info, or an error message.
///
/// Since the input is human-generated, may vary a bit over time, and have optional fields, I've
/// opted to use the generic JSON parser followed by explicit decoding of the fields, rather than a
/// (derived) strongly-typed parser.

pub fn read_cluster_config(filename: &str) -> Result<ClusterConfig> {
    let file = File::open(path::Path::new(filename))?;
    let reader = BufReader::new(file);
    let v = serde_json::from_reader(reader)?;
    let mut m = HashMap::new();
    if let Value::Array(objs) = v {
        for obj in objs {
            if let Value::Object(fields) = obj {
                let mut sys: System = Default::default();
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
                let cross_node_jobs = grab_bool_opt(&fields, "cross_node_jobs")?;
                sys.cross_node_jobs = cross_node_jobs.unwrap_or(false);
                sys.cpu_cores = grab_usize(&fields, "cpu_cores")?;
                sys.mem_gb = grab_usize(&fields, "mem_gb")?;
                let gpu_cards = grab_usize_opt(&fields, "gpu_cards")?;
                let gpumem_gb = grab_usize_opt(&fields, "gpumem_gb")?;
                let gpumem_pct = grab_bool_opt(&fields, "gpumem_pct")?;
                if gpu_cards.is_none() && (gpumem_gb.is_some() || gpumem_pct.is_some()) {
                    bail!("Without gpu_cards there should be no gpumem_gb or gpumem_pct")
                }
                sys.gpu_cards = gpu_cards.unwrap_or(0);
                sys.gpumem_gb = gpumem_gb.unwrap_or(0);
                sys.gpumem_pct = gpumem_pct.unwrap_or(false);
                if let Some(Value::String(ts)) = fields.get("timestamp") {
                    sys.timestamp = ts.clone();
                }
                if let Some(Value::String(ts)) = fields.get("comment") {
                    sys.comment = ts.clone();
                }
                for exp in crate::expand_pattern(&sys.hostname)?.drain(0..) {
                    if m.contains_key(&exp) {
                        bail!("System info for host {exp} already defined");
                    }
                    let key = exp.to_string();
                    m.insert(
                        key,
                        Rc::new(System {
                            hostname: exp,
                            ..sys.clone()
                        }),
                    );
                }
            } else {
                bail!("Expected an object value")
            }
        }
    } else {
        bail!("Expected an array value")
    }
    Ok(ClusterConfig { nodes: m })
}

fn grab_usize(fields: &serde_json::Map<String, Value>, name: &str) -> Result<usize> {
    if let Some(n) = grab_usize_opt(fields, name)? {
        Ok(n)
    } else {
        bail!("Field '{name}' must be present and have an integer value")
    }
}

fn grab_usize_opt(fields: &serde_json::Map<String, Value>, name: &str) -> Result<Option<usize>> {
    if let Some(Value::Number(cores)) = fields.get(name) {
        if let Some(n) = cores.as_u64() {
            match usize::try_from(n) {
                Ok(n) => Ok(Some(n)),
                Err(_e) => {
                    bail!("Field '{name}' must have unsigned integer value")
                }
            }
        } else {
            bail!("Field '{name}' must have unsigned integer value")
        }
    } else {
        Ok(None)
    }
}

fn grab_bool_opt(fields: &serde_json::Map<String, Value>, name: &str) -> Result<Option<bool>> {
    if let Some(d) = fields.get(name) {
        if let Value::Bool(b) = d {
            Ok(Some(*b))
        } else {
            bail!("Field 'gpumem_pct' must have a boolean value");
        }
    } else {
        Ok(None)
    }
}

// Basic whitebox test that the reading works.  Error conditions are tested blackbox, see
// tests/sonalyze/config-file.sh.

#[test]
fn test_config() {
    let conf = read_cluster_config("../tests/sonarlog/whitebox-config.json").unwrap();
    let c0 = conf.lookup("ml1.hpc.uio.no").unwrap();
    let c1 = conf.lookup("ml8.hpc.uio.no").unwrap();
    let c2 = conf.lookup("c1-23").unwrap();
    let c4 = conf.lookup("c1-25").unwrap();

    assert!(&c0.hostname == "ml1.hpc.uio.no");
    assert!(c0.cpu_cores == 56);
    assert!(c0.gpu_cards == 4);
    assert!(c0.gpumem_gb == 0);
    assert!(c0.gpumem_pct == true);

    assert!(&c1.hostname == "ml8.hpc.uio.no");
    assert!(c1.gpu_cards == 3);
    assert!(c1.gpumem_gb == 128);
    assert!(c1.gpumem_pct == false);

    assert!(&c2.hostname == "c1-23");
    assert!(c2.gpu_cards == 0);
    assert!(c2.gpumem_gb == 0);
    assert!(c2.gpumem_pct == false);

    assert!(&c4.hostname == "c1-25");
    assert!(c4.gpu_cards == 0);
    assert!(c4.gpumem_gb == 0);
    assert!(c4.gpumem_pct == false);

    assert!(conf.lookup("ml2.hpc.uio.no").is_none());
}
/// Read system configuration data for a cluster from a json file.
///
/// See ../../../production/sonar-nodes/$CLUSTER/$CLUSTER-config.json for examples.
///
/// v2 file format:
///
/// An object { ... } with the following named fields and value types:
///
///   name - string, the canonical name of the cluster
///   description - string, optional, arbitrary text describing the cluster
///   aliases - array of strings, optional, aliases / short names for the cluster
///   exclude-user - array of strings, optional, user names whose records should
///      be excluded when filtering records
///   nodes - array of objects, the list of nodes in the v1 format (see below)
///
/// Any field name starting with '#' is reserved for arbitrary comments.
///
/// The `exclude-user` option is a hack and is used to add post-hoc filtering of data (when Sonar
/// should have filtered it to begin with, but didn't).  It is on purpose very limited, in contrast
/// with e.g. a mechanism to add arbitrary arguments to the command line.  Additional filters, eg
/// for command names, can be added as needed.
///
/// v1 file format:
///
/// An array [...] of objects { ... }, each with the following named fields and value types:
///
///   timestamp - string, optional, an RFC3339 timestamp for when the data were obtained
///   hostname - string, the fully qualified and unique host name of the node
///   description - string, optional, arbitrary text describing the node
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
/// Any field name starting with '#' is reserved for arbitrary comments.
///
use crate::Timestamp;
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
    pub cross_node_jobs: bool,
    pub cpu_cores: usize,
    pub mem_gb: usize,
    pub gpu_cards: usize,
    pub gpumem_gb: usize,
    pub gpumem_pct: bool,
}

#[derive(Default)]
pub struct ClusterConfig {
    pub name: String,
    pub description: String,
    pub aliases: Vec<String>,
    pub exclude_user: Vec<String>,
    nodes: HashMap<String, Rc<System>>,
}

impl ClusterConfig {
    pub fn lookup(&self, hostname: &str) -> Option<Rc<System>> {
        self.nodes.get(hostname).cloned()
    }

    // Returns the hosts that were defined within the time window.  With our current structure we don't
    // have reliable time window information, so just return all hosts.
    pub fn hosts_in_time_window(&self, _from_incl: Timestamp, _to_excl: Timestamp) -> Vec<String> {
        self.nodes.iter().map(|(k,_)| k.to_string()).collect::<Vec<String>>()
    }

    pub fn cross_node_jobs(&self) -> bool {
        self.nodes.iter().any(|(_, sys)| sys.cross_node_jobs)
    }
}

/// Since the input is human-generated, may vary a lot over time, and have optional fields, I've
/// opted to use the generic JSON parser followed by explicit decoding of the fields, rather than a
/// (derived) strongly-typed parser.

pub fn read_cluster_config(filename: &str) -> Result<ClusterConfig> {
    let file = File::open(path::Path::new(filename))?;
    let reader = BufReader::new(file);
    let v = serde_json::from_reader(reader)?;
    let mut cfg: ClusterConfig = Default::default();
    if let Value::Array(objs) = v {
        cfg.name = filename.to_string();
        cfg.description = filename.to_string();
        cfg.nodes = process_cluster_nodes(&objs)?;
        // Aliases could be set to be the first element of the host name of the config file name,
        // this would usually be right.
    } else if let Value::Object(fields) = v {
        cfg.name = grab_string(&fields, "name")?;
        cfg.description = grab_string_opt(&fields, "description")?;
        cfg.aliases = grab_strings_opt(&fields, "aliases")?;
        cfg.exclude_user = grab_strings_opt(&fields, "exclude-user")?;
        if let Some(Value::Array(objs)) = fields.get("nodes") {
            cfg.nodes = process_cluster_nodes(objs)?;
        } else {
            bail!("The field 'nodes' is required");
        }
    } else {
        bail!("Expected an array value")
    }
    Ok(cfg)
}

fn process_cluster_nodes(objs: &Vec<Value>) -> Result<HashMap<String, Rc<System>>> {
    let mut nodes = HashMap::new();
    for obj in objs {
        if let Value::Object(fields) = obj {
            let mut sys: System = Default::default();
            sys.hostname = grab_string(&fields, "hostname")?;
            sys.description = grab_string_opt(&fields, "description")?;
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
            sys.timestamp = grab_string_opt(&fields, "timestamp")?;
            for exp in crate::expand_pattern(&sys.hostname)?.drain(0..) {
                if nodes.contains_key(&exp) {
                    bail!("System info for host {exp} already defined");
                }
                let key = exp.to_string();
                nodes.insert(
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
    Ok(nodes)
}

fn grab_string(fields: &serde_json::Map<String, Value>, name: &str) -> Result<String> {
    if let Some(Value::String(s)) = fields.get(name) {
        Ok(s.to_string())
    } else {
        bail!("Field '{name}' must be present and have a string value");
    }
}

fn grab_string_opt(fields: &serde_json::Map<String, Value>, name: &str) -> Result<String> {
    if let Some(val) = fields.get(name) {
        if let Value::String(s) = val {
            Ok(s.to_string())
        } else {
            bail!("Field '{name}' must have a string value");
        }
    } else {
        Ok("".to_string())
    }
}

fn grab_strings_opt(fields: &serde_json::Map<String, Value>, name: &str) -> Result<Vec<String>> {
    let mut result = vec![];
    if let Some(Value::Array(vals)) = fields.get(name) {
        for v in vals {
            if let Value::String(s) = v {
                result.push(s.to_string());
            } else {
                bail!("Field '{name}' must have string values");
            }
        }
    }
    Ok(result)
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

// Basic whitebox test that the reading of v1 configs works.  Error conditions are tested blackbox,
// see tests/config/config-file.sh.

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

// Basic whitebox test that the reading of v2 configs works.  Error conditions are tested blackbox,
// see tests/config/config-file.sh.

#[test]
fn test_config_v2() {
    let conf = read_cluster_config("../tests/sonarlog/whitebox-config-v2.json").unwrap();
    assert!(conf.name == "mlx.hpc.uio.no");
    assert!(conf.description == "UiO machine learning nodes");
    assert!(conf.aliases.len() == 2 && conf.aliases[0] == "ml" && conf.aliases[1] == "mlx");
    assert!(conf.exclude_user.len() == 2 && conf.exclude_user[0] == "root" && conf.exclude_user[1] == "toor");

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

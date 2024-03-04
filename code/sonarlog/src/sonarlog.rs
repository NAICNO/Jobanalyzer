use crate::{InputStreamSet, LogEntry};

use anyhow::Result;

pub struct Sonarlog {
    files: Option<Vec<String>>,
    dir: Option<String>,
}

impl Sonarlog {
    pub fn open_dir(path: &str) -> Result<Sonarlog> {
        Ok(Sonarlog {
            files: None,
            dir: Some(path.to_string()),
        })
    }

    pub fn open_files(files: &[String]) -> Result<Sonarlog> {
        Ok(Sonarlog {
            files: files.iter().map(|x| x.to_string()).collect::Vec<String>(),
            dir: None,
        })
    }

    pub fn store_events(&mut self, evs: Vec<Box<LogEntry>>) -> Result<()> {
        // must have dir, not files
        // append events to right files
    }

    // Probably wants filter and config both, config optional
    pub fn get_streams(&self, filter: F) -> Result<InputStreamSet> {
        let events = self.read()?;
        ...;
    }

    // Ditto
    pub fn get_events(&self, filter: F) -> Result<Vec<Box<LogEntry>>> {
        let events = self.read()?;
        logfile::basic_cleaning(events)?;
        ...;
    }

    // Ditto
    fn read(&self) -> Result<Vec<Box<LogEntry>>> {
        // Read records and return raw events
    }
}

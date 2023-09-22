/// Qua the pattern grammar in hosts.rs, `expand_element` syntax checks and expands the
/// number ranges of an "element" nonterminal and returns the vector of the expanded strings.  Each
/// string may still be suffixed by '*'.

use anyhow::{bail, Result};
use std::str::FromStr;

pub fn expand_element(s: &str) -> Result<Vec<String>> {
    let mut parser = Parser::new(s);
    parser.parse_nonempty_element()?;
    Ok(parser.result)
}

struct Parser {
    result: Vec<String>, // Accumulated pattern strings
    ss: Vec<String>,     // Current set of strings for eltpat
    input: Vec<char>,    // Vector of input characters
    i: usize,            // Index into input
    lim: usize,          // Length of input
}

impl Parser {
    fn new(s: &str) -> Parser {
        let input = s.chars().collect::<Vec<char>>();
        let lim = input.len();
        Parser {
            result: vec![],
            ss: vec!["".to_string()],
            input,
            i: 0,
            lim,
        }
    }

    fn parse_nonempty_element(&mut self) -> Result<()> {
        self.parse_nonempty_eltpat()?;
        while !self.at_end() {
            self.match_char(',')?;
            self.parse_nonempty_eltpat()?;
        }
        Ok(())
    }

    // This will not consume the ',' following the eltpat.

    fn parse_nonempty_eltpat(&mut self) -> Result<()> {
        'eltpat: loop {
            if self.at_end() {
                self.consume_nonempty()?;
                break 'eltpat;
            }
            match self.peek()? {
                ',' => {
                    self.consume_nonempty()?;
                    break 'eltpat;
                }
                '*' => {
                    self.next();
                    self.push_char('*');
                    self.consume_nonempty()?;
                    break 'eltpat;
                }
                '[' => {
                    self.next();
                    let numbers = self.parse_brackets()?; // Consumes ']'
                    self.push_numbers(numbers);
                    continue 'eltpat;
                }
                c => {
                    self.next();
                    self.push_char(c);
                }
            }
        }
        Ok(())
    }

    // The '[' has been eaten, and this consumes the ']' but does not look beyond that.

    fn parse_brackets(&mut self) -> Result<Vec<String>> {
        let mut lst = vec![];
        lst.extend(self.parse_range()?);
        while self.peek()? != ']' {
            self.match_char(',')?;
            lst.extend(self.parse_range()?);
        }
        self.match_char(']')?;
        Ok(lst)
    }

    // Consumes either m or m-n, peeks at the next input element.

    fn parse_range(&mut self) -> Result<Vec<String>> {
        let mut result = vec![];
        let m = self.parse_u32()?;
        if self.peek()? == '-' {
            self.next();
            let n = self.parse_u32()?;
            let mut i = m;
            while i <= n {
                result.push(i.to_string());
                i += 1;
            }
        } else {
            result.push(m.to_string())
        }
        Ok(result)
    }

    // Consumes digits, errors out on an empty string, peeks at the next input element.

    fn parse_u32(&mut self) -> Result<u32> {
        let start = self.i;
        let mut s = "".to_string();
        while self.i < self.lim && self.input[self.i].is_ascii_digit() {
            s.push(self.input[self.i]);
            self.i += 1;
        }
        if self.i == start {
            bail!("Expected number");
        }
        Ok(u32::from_str(&s)?)
    }

    // Result accumulation abstraction.

    fn push_char(&mut self, c: char) {
        for s in &mut self.ss {
            s.push(c);
        }
    }

    fn push_numbers(&mut self, ns: Vec<String>) {
        let mut nvec = vec![];
        for s in &mut self.ss {
            for n in &ns {
                let mut x = s.clone();
                x += n;
                nvec.push(x);
            }
        }
        self.ss = nvec;
    }

    fn consume_nonempty(&mut self) -> Result<()> {
        if self.ss.len() == 1 && self.ss[0].len() == 0 {
            bail!("Empty pattern in input")
        }
        self.result.extend_from_slice(&self.ss);
        self.ss = vec!["".to_string()];
        Ok(())
    }

    // Input stream abstraction

    fn match_char(&mut self, c: char) -> Result<()> {
        if self.get()? != c {
            bail!("Expected {c}");
        }
        Ok(())
    }

    fn get(&mut self) -> Result<char> {
        let c = self.peek()?;
        self.next();
        Ok(c)
    }

    fn peek(&mut self) -> Result<char> {
        if self.at_end() {
            bail!("Unexpected end of input")
        }
        let c = self.input[self.i];
        Ok(c)
    }

    fn next(&mut self) {
        if self.at_end() {
            panic!("Unexpected EOI");
        }
        self.i += 1;
    }

    fn at_end(&self) -> bool {
        self.i == self.lim
    }
}

// Tests copied from https://github.com/NordicHPC/jobgraph (src/nodelist.rs) and modified some.

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_expand_element() {
        assert_eq!(
            expand_element("c1-[0-1],c2-[2-3]").unwrap(),
            vec!["c1-0", "c1-1", "c2-2", "c2-3"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c1-0,c2-0").unwrap(),
            vec!["c1-0", "c2-0"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c1-0,c2-1").unwrap(),
            vec!["c1-0", "c2-1"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c2-1").unwrap(),
            vec!["c2-1"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c2-[1,3,5]").unwrap(),
            vec!["c2-1", "c2-3", "c2-5"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c2-[1-3,5]").unwrap(),
            vec!["c2-1", "c2-2", "c2-3", "c2-5"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c3-[1-3,5,9-12]").unwrap(),
            vec!["c3-1", "c3-2", "c3-3", "c3-5", "c3-9", "c3-10", "c3-11", "c3-12"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c3-[5,9-12]").unwrap(),
            vec!["c3-5", "c3-9", "c3-10", "c3-11", "c3-12"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c3-[5,9],c5-[15-19]").unwrap(),
            vec!["c3-5", "c3-9", "c5-15", "c5-16", "c5-17", "c5-18", "c5-19"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c3-[5,9],c5-[15,17]").unwrap(),
            vec!["c3-5", "c3-9", "c5-15", "c5-17"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c3-5,c7-[15,17]").unwrap(),
            vec!["c3-5", "c7-15", "c7-17"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c3-[5,9],c8-175").unwrap(),
            vec!["c3-5", "c3-9", "c8-175"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c1-20").unwrap(),
            vec!["c1-20"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c1-34,c2-[3,21]").unwrap(),
            vec!["c1-34", "c2-3", "c2-21"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c1-[34,37-38,41]").unwrap(),
            vec!["c1-34", "c1-37", "c1-38", "c1-41"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c5-54,c11-30").unwrap(),
            vec!["c5-54", "c11-30"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );

        assert_eq!(
            expand_element("c2-[1,3-5]").unwrap(),
            vec!["c2-1", "c2-3", "c2-4", "c2-5"]
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>()
        );
    }
}

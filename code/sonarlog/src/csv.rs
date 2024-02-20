/// Non-allocating CSV tokenizer.
///
/// CSV is not a single well-defined format.  Here is what we're parsing:
///
///  - the input is UTF-8 / ASCII, no BOM allowed (or needed).
///  - there is one CSV record per line
///  - lines are terminated by ASCII newline 0x0A, exclusively
///  - the terminator is optional at EOF
///  - blank lines are treated as empty records
///  - lines have a hardcoded max length given by MAXLINE, below
///  - there is no header line
///  - fields are separated by ASCII comma 0x2C, exclusively
///  - the number of fields can vary between the lines
///  - fields can be empty
///  - fields can be enclosed in double-quotes ASCII 0x22 and double-quotes and commas are allowed
///    inside quoted fields
///  - newlines and EOF are not allowed inside double-quoted fields
///  - a double-quote is represented as two double-quotes in a double-quoted field
///
/// The tokenizer is opened on an io::Read and provides tokens.  A token is one of Tok::Field,
/// Tok::EOL, and Tok::EOF.  The Tok::Field is a triplet: start index into byte buffer, one-past-end
/// index into byte buffer, and index within that range of the character following the first '=', if
/// present, otherwise EQ_SENTINEL.  The indices are valid until the next call to get() and can be
/// passed to buf_at(), get_str() and get_string() to retrieve contents from the internal buffer.
///
/// This is much more efficient than the serde-derived CSV parser, since the best we could do with
/// that - given the flexibility of the input format - was to parse the record as a vector of string
/// fields.  The allocation volume was tremendous.
use anyhow::{bail, Result};
use std::io;

pub enum Token {
    Field(usize, usize, usize),
    EOL,
    EOF,
}

pub const EQ_SENTINEL: usize = usize::MAX;

const BUFSIZ: usize = 65536;
const MAXLINE: usize = 1024;

pub struct Tokenizer<'a> {
    ix: usize,
    lim: usize,
    start_of_line: bool,
    reader: &'a mut dyn io::Read,
    buf: [u8; BUFSIZ],
    #[cfg(test)]
    fail_refill: bool,
}

impl<'a> Tokenizer<'a> {
    pub fn new(reader: &'a mut dyn io::Read) -> Box<Tokenizer> {
        Box::new(Tokenizer {
            ix: 0,
            lim: 0,
            start_of_line: true,
            reader: reader,
            buf: [0u8; BUFSIZ],
            #[cfg(test)]
            fail_refill: false,
        })
    }

    // Extract a &str reference into the buffer.  start and lim must have been returned with a
    // Token::Field from get().  This &str is only valid until the next call to get().
    pub fn get_str(&self, start: usize, lim: usize) -> &str {
        unsafe { std::str::from_utf8_unchecked(&self.buf[start..lim]) }
    }

    // Get the byte in the buffer at the given locations, this is valid exclusively for locations
    // start..lim returned by a Token::Field until the next call to get().
    pub fn buf_at(&self, loc: usize) -> u8 {
        self.buf[loc]
    }

    // Get the next token, or an error for syntax errors or I/O errors.
    pub fn get(&mut self) -> Result<Token> {
        self.maybe_refill()?;

        // The following logic assumes \n is a sentinel at self.buf[self.lim].

        if self.buf[self.ix] == b'\n' {
            if self.ix == self.lim {
                return Ok(Token::EOF);
            }
            self.ix += 1;
            self.start_of_line = true;
            return Ok(Token::EOL);
        }

        if !self.start_of_line {
            assert!(self.buf[self.ix] == b',');
            self.ix += 1;
        }

        // Parse a field, terminated by EOF, EOL, or comma.
        let mut eqloc = EQ_SENTINEL;
        self.start_of_line = false;
        match self.buf[self.ix] {
            b'\n' | b',' => {
                // Empty field at EOL or EOF or comma
                return Ok(Token::Field(self.ix, self.ix, eqloc));
            }
            b'"' => {
                // This is a little hairy because every doubled quote has to be collapsed into a
                // single one.  We do this in the buffer.
                self.ix += 1;
                let startix = self.ix;
                let mut destix = startix;
                loop {
                    match self.buf[self.ix] {
                        b'\n' => {
                            bail!("Unexpected end of line or end of file");
                        }
                        b'=' => {
                            if eqloc == EQ_SENTINEL {
                                eqloc = self.ix + 1
                            }
                        }
                        b'"' => {
                            self.ix += 1;
                            if self.buf[self.ix] != b'"' {
                                // We're done.  We've already consumed the quote.  Check that the
                                // syntax is sane.
                                if self.buf[self.ix] != b',' && self.buf[self.ix] != b'\n' {
                                    bail!("Expected comma or newline after quoted field")
                                }
                                return Ok(Token::Field(startix, destix, eqloc));
                            }
                        }
                        _ => {}
                    }
                    self.buf[destix] = self.buf[self.ix];
                    destix += 1;
                    self.ix += 1;
                }
            }
            _ => {
                let startix = self.ix;
                loop {
                    match self.buf[self.ix] {
                        b'\n' | b',' => {
                            return Ok(Token::Field(startix, self.ix, eqloc));
                        }
                        b'=' => {
                            if eqloc == EQ_SENTINEL {
                                eqloc = self.ix + 1
                            }
                        }
                        b'"' => {
                            bail!("Unexpected '\"'");
                        }
                        _ => {}
                    }
                    self.ix += 1;
                }
            }
        }
    }

    // Given start and non-sentinel eqloc values returned with a Token::Field, and a string <tag>,
    // check if the buffer has the string <tag>= from location start.
    pub fn match_tag(&self, tag: &[u8], start: usize, eqloc: usize) -> bool {
        if start + tag.len() + 1 != eqloc {
            return false;
        }
        if self.buf[eqloc - 1] != b'=' {
            return false;
        }
        for i in 0..tag.len() {
            if tag[i] != self.buf[start + i] {
                return false;
            }
        }
        return true;
    }

    fn maybe_refill(&mut self) -> Result<()> {
        while self.lim - self.ix < MAXLINE {
            if self.ix != 0 {
                let n = self.lim - self.ix;
                self.buf.copy_within(self.ix..self.lim + 1, 0);
                self.ix = 0;
                self.lim = n;
            }
            #[cfg(test)]
            if self.fail_refill {
                return Err(io::Error::new(io::ErrorKind::Unsupported, "Test failure").into());
            }
            let nread = self.reader.read(&mut self.buf[self.lim..BUFSIZ - 1])?;
            self.lim += nread;
            self.buf[self.lim] = b'\n';
            if nread == 0 {
                break;
            }
        }
        Ok(())
    }

    #[cfg(test)]
    fn set_fail_refill(&mut self) {
        self.fail_refill = true;
    }
}

// This tests:
//  - empty fields, also at EOL (but not at EOF)
//  - quoted fields, also with commas and quotes in them
//  - unterminated last line
//  - blank line / empty record

#[test]
fn test_csv_tokenizer1() {
    let text = r#"a,b=1,cc=2,,e,"f=1,2,3","g,""y"",z",

A,B"#;
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "a");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "b=1");
        assert!(c == a + 2);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "cc=2");
        assert!(c == a + 3);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "e");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "f=1,2,3");
        assert!(c == a + 2);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "g,\"y\",z");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::EOL = tokenizer.get().unwrap() {
    } else {
        assert!(false);
    }
    if let Token::EOL = tokenizer.get().unwrap() {
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "A");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "B");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::EOF = tokenizer.get().unwrap() {
    } else {
        assert!(false);
    }
}

// This tests:
//  - empty field at EOF

#[test]
fn test_csv_tokenizer2() {
    let text = r#"a,"#;
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "a");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    if let Token::EOF = tokenizer.get().unwrap() {
    } else {
        assert!(false);
    }
}

// This tests:
//  - syntax error: eol in quoted string

#[test]
fn test_csv_tokenizer3() {
    let text = r#"a,"hi
ho"#;
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "a");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    assert!(tokenizer.get().is_err());
}

// This tests:
//  - syntax error: eof in quoted string

#[test]
fn test_csv_tokenizer4() {
    let text = r#"a,"hi"#;
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "a");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    assert!(tokenizer.get().is_err());
}

// This tests:
//  - syntax error: junk following quoted string

#[test]
fn test_csv_tokenizer5() {
    let text = r#"a,"hi"x,y"#;
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "a");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    assert!(tokenizer.get().is_err());
}

// This tests:
//  - syntax error: quote in unquoted string

#[test]
fn test_csv_tokenizer6() {
    let text = r#"a,hi"x,y"#;
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    if let Token::Field(a, b, c) = tokenizer.get().unwrap() {
        assert!(tokenizer.get_str(a, b) == "a");
        assert!(c == EQ_SENTINEL);
    } else {
        assert!(false);
    }
    assert!(tokenizer.get().is_err());
}

// This tests:
//  - refill logic
//
// Basically we're creating an input s.t. the characters for a token overlap the buffer boundary.
// Parts of the token are read on the initial fill, then the rest on the second fill.  The file
// contains single-token lines, each line is the 26-character string a...z followed by a newline.
// The test then just checks that every token looks right.

#[test]
fn test_csv_tokenizer7() {
    // Token+newline must straddle the buffer boundary in some interesting way
    assert!(BUFSIZ % 27 != 0 && BUFSIZ % 27 != 1 && BUFSIZ % 27 != 26);
    let mut text = "".to_string();
    let count = BUFSIZ * 3 / 27;
    for _i in 0..count {
        text += "abcdefghijklmnopqrstuvwxyz\n";
    }
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    let mut state = 0;
    let mut found = 0;
    loop {
        match tokenizer.get().unwrap() {
            Token::Field(a, b, c) => {
                assert!(state == 0);
                assert!(tokenizer.get_str(a, b) == "abcdefghijklmnopqrstuvwxyz");
                assert!(c == EQ_SENTINEL);
                state = 1;
                found += 1;
            }
            Token::EOL => {
                assert!(state == 1);
                state = 0;
            }
            Token::EOF => {
                assert!(state == 0);
                break;
            }
        }
    }
    assert!(found == count);
}

// This tests:
//  - i/o error on refill

#[test]
fn test_csv_tokenizer8() {
    // This is the same input as test_csv_tokenizer7() for a reason, see below.
    let mut text = "".to_string();
    let count = BUFSIZ * 3 / 27;
    for _i in 0..count {
        text += "abcdefghijklmnopqrstuvwxyz\n";
    }
    let mut bs = text.as_bytes();
    let mut tokenizer = Tokenizer::new(&mut bs);
    tokenizer.get().unwrap(); // Fill once
    tokenizer.set_fail_refill();
    loop {
        match tokenizer.get() {
            Ok(Token::EOF) => {
                assert!(false);
            }
            Ok(_) => {}
            Err(_e) => {
                // This can only be an I/O error because test_csv_tokenizer7() would otherwise have
                // encountered it, since it uses the same input.
                break;
            }
        }
    }
}

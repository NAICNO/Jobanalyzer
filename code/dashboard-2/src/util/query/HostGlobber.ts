// This is an implementation of a host name *matcher* as defined by the documentation at the head of
// ../go-utils/hostglob/hostglob.go.  See that for a full description of the grammar and semantics.
//
// Note, there is also a Rust implementation of matching in sonarlog/src/hostglob.rs; it must be
// kept in sync with this one.  Both must be kept in sync with the specification in the Go code.
//
// See HostGlobber.test.ts for test cases.

// This was rewritten in TypeScript from the original JS code. The original JS code was written by @Lars T Hansen.

// This takes a <multi-pattern> according to the grammar referenced above and returns a list of
// individual <pattern>s in that list.  It requires a bit of logic because each pattern may contain
// a fragment that contains a comma.  It throws an error if it encounters bad syntax.

export function splitMultiPattern(s: string) {
  const strings: string[] = []
  if (s === '') {
    return strings
  }
  let insideBrackets = false
  let start = -1
  for (let ix = 0; ix < s.length; ix++) {
    const c = s.charAt(ix)
    if (c == '[') {
      if (insideBrackets) {
        throw new Error('Illegal pattern: nested brackets')
      }
      insideBrackets = true
    } else if (c == ']') {
      if (!insideBrackets) {
        throw new Error('Illegal pattern: unmatched end bracket')
      }
      insideBrackets = false
    } else if (c == ',' && !insideBrackets) {
      if (start == -1) {
        throw new Error('Illegal pattern: Empty host name')
      }
      strings.push(s.substring(start, ix))
      start = -1
    } else if (start == -1) {
      start = ix
    }
  }
  if (insideBrackets) {
    throw new Error('Illegal pattern: Missing end bracket')
  }
  if (start == s.length || start == -1) {
    throw new Error('Illegal pattern: Empty host name')
  }
  strings.push(s.substring(start))
  return strings
}

// Given a <pattern> expression in the grammar referenced above, `new HostGlobber(pattern)` returns
// a matcher object with a method `match(hostname)` which returns true iff the hostname is matched
// exactly by the pattern (the number of pattern-elements must equal the number of host-elements and
// each must match precisely).
//
// In the event of an error in the pattern, this throws an Error.

export class HostGlobber {
  private readonly re: RegExp

  constructor(pattern: string, prefix?: boolean) {
    // @ts-ignore
    const [re, reSrc] = this.compileGlobber(pattern, prefix)
    this.re = re
  }

  // Translate the <pattern> p into a regular expression.  Return the compiled regex and the source
  // for it.  Throws an error for every error.

  private compileGlobber(p: string, prefix?: boolean): [RegExp, string] {
    let i = 0
    let r = '^'

    const isDigit = (c: string): boolean => {
      return c >= '0' && c <= '9'
    }

    const readInt = (): number => {
      let s = ''
      for (; isDigit(p.charAt(i)); i++) {
        s += p.charAt(i)
      }
      if (s === '') {
        throw new Error('Invalid number in glob set')
      }
      let n = parseInt(s, 10)
      if (n > 0xFFFFFFFF) {
        throw new Error('Number out of range in glob set')
      }
      return n
    }

    // We don't need to check the string limit, charAt returns "" past the end.

    Loop:
      for (; ;) {
        if (r.length > 50000) {
          // Avoid pathological behavior
          throw new Error('Expression too large, use more \'*\'')
        }
        switch (p.charAt(i)) {
        case '':
          break Loop
        case '*':
          i++
          r += '[^.]*'
          break
        case '[':
          i++
          // Single chars or ranges, separated by commas.
          //
          // Note, it's possible to compress many ranges.  For example, 1-5 is just the charset
          // [1-5], it doesn't need to be (?:1|2|3|4|5).  The range [10-19] is 1[0-9], i.e. 1\d.
          // [10-20] is 1\d|20, [10-30] is 1\d|2\d|30.  In practice, we don't care about that
          // level of efficiency.
          let set: string[] = []
          for (; ;) {
            let n0 = readInt()
            if (p.charAt(i) === '-') {
              i++
              let n1 = readInt()
              if (n0 > n1) {
                throw new Error('Invalid range')
              }
              for (; n0 <= n1; n0++) {
                set.push(String(n0))
                if (set.length > 10000) {
                  // Avoid pathological behavior, [1-1000000] is a thing.
                  throw new Error('Range too large, use more \'*\'')
                }
              }
            } else {
              set.push(String(n0))
            }
            if (p.charAt(i) === ']') {
              i++
              break
            }
            if (p.charAt(i) !== ',') {
              throw new Error('Expected \',\'')
            }
            i++
          }
          r += '(?:' + set.join('|') + ')'
          break
        case '.':
        case '$':
        case '^':
        case '?':
        case '\\':
        case '(':
        case ')':
        case ']':
        case '{':
        case '}':
          // Mostly these special chars are not allowed in hostnames, but it doesn't hurt to
          // try to avoid disasters.
          r += '\\'
          r += p.charAt(i)
          i++
          break
        case ',':
          throw new Error('\',\' not allowed here')
        default:
          r += p.charAt(i)
          i++
        }
      }
    if (prefix) {
      // Matching a prefix: we need to match entire host-elements, so following a prefix there
      // should either be EOS or `.` followed by whatever until EOS.
      r += '(?:\\..*)?$'
    } else {
      r += '$'
    }
    return [new RegExp(r), r]
  }

  match(hn: string): boolean {
    hn = String(hn)
    return hn.match(this.re) !== null
  }
}

// General-ish query engine for node data.
//
// The purpose of this is to allow large and undisplayable sets of data rows (typically one row per
// node but it could also be used to query jobs) to be culled to something more manageable.
//
// It works by compiling a simple boolean/relational query under an environment into a query matcher
// object, that can then be applied to data rows.
//
// The query matcher takes as its input an immutable table of data objects (rows) and a
// representation of the set of rows (indices) from that set that are currently considered.  It
// returns a new set, the result of the filter, a subset of the input set.  (It is based on a set of
// rows rather than a single row because culling by eg node names in terms evaluated early will tend
// to quickly reduce the number of data rows considered by terms evaluated later, but that
// optimization is not currently implemented.)
//
// For example, the primitive query "(mem% > 50)" constructs a set that comprises those elements in
// the input set whose mem% value is greater than 50.  For another example, the primitive query
// "c1-*" constructs a set that comprises those elements in the input set whose node names match
// that pattern.  The query "c1-* and mem% > 50" combines them.  "login and down" shows the login
// nodes that are currently believed to be down.
//
// Note that the actual defined variable terms ("mem%" etc) and predefined operators for node
// selection ("login", "down", etc) are defined by the client of this library.
//
// For test cases, see QueryUtils.test.ts

// Dependencies:
//   "Bitset.ts"
//   "HostGlobber.ts"

// Query compiler.
//
// The `query` is the query expression.  The `knownFields` is a map from known field names to either
// `true` or to a canonical field name (allowing for aliases).  The `builtinOperations` is a map
// from expression aliases (essentially subroutines) to their matcher expandings.
//
// Returns either a query matcher object, or a string describing the error as precisely as possible.
//
// The expression grammar is simple:
//
//   expr ::= token binop token
//          | unop token
//          | token
//          | (expr)
//
//   binop ::= "and" | "or" | "<" | ">" | "<=" | ">=" | "="
//   unop ::= "~"
//   token ::= string of non-whitespace chars other than punctuation, except "and" and "or"
//   punctuation ::= "<", ">", "=", "~", "(", ")"
//
// Everything is case-sensitive.
//
// Tokens represent either known field references, numbers, or node name wildcard matchers.  The
// interpretation of a token is contextual: a field name or number is interpreted as such only in
// the context of a binary operator that requires that interpretation; in all other contexts they
// are interpreted as node name matchers.  That is, "c1* > 5 and 37.5" is a legal and meaningful
// instruction if there is a known field called "c1*" and a node called "37.5".
//
// Binary operator precedence from low to high: OR, AND, relationals.  Unary ops bind tighter
// than binary ops.
//
// For a relational operator, the first argument must be a field name and the second must be a
// number.
//
// The meaning of an expression is that each node name matcher or relational operation induces a
// subset of the data rows, "and" is set intersection, "or" is set union, and "~" is set complement
// (where the full set of nodes is the universe).

import { NotOperation } from './NotOperation.ts'
import { RelOperation } from './RelOperation.ts'
import { SetOperation } from './SetOperation.ts'
import { GlobOperation } from './GlobOperation.ts'
import { Bitset } from './Bitset.ts'
import JobQueryValues from '../../types/JobQueryValues.ts'
import { splitMultiPattern } from './HostGlobber.ts'
import { TRUE_VAL } from '../../Constants.ts'

export function compileQuery(query: string, knownFields: Record<string, string | boolean> = {}, builtinOperations: Record<string, any> = {}) {

  // Character indices in `query`
  let i = 0
  const lim = query.length

  // Location of last token gotten, used for error reporting.
  let loc = 0

  // Token stream, with start-of-token location, -1 means "no token yet"
  let pending = ''
  let pending_i = -1

  const operatorNames = {
    '<': true, '<=': true, '>': true, '>=': true, '=': true, 'and': true, 'or': true, '~': true,
  }

  // Place a token in pending if not yet there, return false if no more tokens could be extracted.
  function fill() {
    if (pending_i >= 0) {
      return true
    }
    while (i < lim && isSpace(query.charAt(i))) {
      i++
    }
    if (i == lim) {
      return false
    }
    pending_i = i
    const probe = query.charAt(i)
    if (isDelim(probe)) {
      pending = probe
      i++
      return true
    }
    let isOperator = false
    if (isPunct(probe)) {
      i++
      while (i < lim && isPunct(query.charAt(i))) {
        i++
      }
      isOperator = true
    } else if (isToken(probe)) {
      i++
      while (i < lim && isToken(query.charAt(i))) {
        i++
      }
    } else {
      fail(`Unexpected character '${probe}'`)
    }
    pending = query.substring(pending_i, i)
    if (isOperator && !(pending in operatorNames)) {
      fail(`Unknown operator '${pending}'`)
    }
    return true
  }

  function get() {
    if (!fill()) {
      loc = i
      fail('Unexpected end of expression')
    }
    const s = pending
    loc = pending_i
    pending_i = -1
    return s
  }

  function fail(irritant: unknown) {
    throw new Error(`Location ${loc + 1}: ${irritant}`)
  }

  function eatToken(t: string) {
    if (!fill()) {
      return false
    }
    if (pending == t) {
      get()
      return true
    }
    return false
  }

  function exprBin(next: any, constructor: any, ...ops : any[]) {
    return function () {
      let e = next()
      Outer:
      for (; ;) {
        for (const op of ops) {
          if (eatToken(op)) {
            const e2 = next()
            try {
              e = new constructor(op, e, e2)
            } catch (ex) {
              fail(ex)
            }
            continue Outer
          }
        }
        break
      }
      return e
    }
  }

  type QueryExpression = NotOperation | RelOperation | SetOperation | GlobOperation | number | string

  function exprPrim() : QueryExpression {
    if (eatToken('(')) {
      const e = exprOr()
      if (!eatToken(')')) {
        fail('Expected \')\' here')
      }
      return e
    }
    if (eatToken('~')) {
      const e : QueryExpression = exprPrim()
      return new NotOperation(e)
    }
    let t = get()
    const probe = parseFloat(t)
    if (isFinite(probe)) {
      return probe
    }
    while (knownFields.hasOwnProperty(t)) {
      const mapping = knownFields[t]
      if (mapping === true)
        return t
      if (typeof mapping != 'string') {
        // something strange
        break
      }
      // alias
      t = mapping
    }
    if (t in builtinOperations) {
      return builtinOperations[t]
    }
    if (t in operatorNames || t == '(' || t == ')') {
      fail(`Misplaced operator or punctuation '${t}'`)
    }
    return new GlobOperation(t)
  }

  const exprRelop = exprBin(exprPrim, RelOperation, '<', '<=', '>', '>=', '=')
  const exprAnd = exprBin(exprRelop, SetOperation, 'and')
  const exprOr = exprBin(exprAnd, SetOperation, 'or')

  function expr() {
    const e = exprOr()
    if (fill()) {
      fail(`Junk at end of expression: ${get()}`)
    }
    return e
  }

  try {
    return expr()
  } catch (ex) {
    return String(ex)
  }
}


const knownFields: Record<string, string | boolean> = {
  // Canonical, these have historical names
  'cpu_recent': true,
  'cpu_longer': true,
  'mem_recent': true,
  'mem_longer': true,
  'resident_recent': true,
  'resident_longer': true,
  'gpu_recent': true,
  'gpu_longer': true,
  'gpumem_recent': true,
  'gpumem_longer': true,
  'cpu_status': true,
  'gpu_status': true,
  'users_recent': true,
  'users_longer': true,
  'jobs_recent': true,
  'jobs_longer': true,
  'violators': true,
  'zombies': true,

  // More sensible naming
  'virt_recent': 'mem_recent',
  'virt_longer': 'mem_longer',
  'res_recent': 'resident_recent',
  'res_longer': 'resident_longer',

  // Dash rather than underscore for user-friendly naming
  'cpu-recent': 'cpu_recent',
  'cpu-longer': 'cpu_longer',
  'virt-recent': 'virt_recent',
  'virt-longer': 'virt_longer',
  'res-recent': 'res_recent',
  'res-longer': 'res_longer',
  'gpu-recent': 'gpu_recent',
  'gpu-longer': 'gpu_longer',
  'gpumem-recent': 'gpumem_recent',
  'gpumem-longer': 'gpumem_longer',
  'cpu-status': 'cpu_status',
  'gpu-status': 'gpu_status',
  'users-recent': 'users_recent',
  'users-longer': 'users_longer',
  'jobs-recent': 'jobs_recent',
  'jobs-longer': 'jobs_longer',

  // Abbreviations - names are case-sensitive, use the Capitalized name for the "longer" field.
  'cpu%': 'cpu-recent',
  'Cpu%': 'cpu-longer',
  'virt%': 'virt-recent',
  'Virt%': 'virt-longer',
  'res%': 'res-recent',
  'Res%': 'res-longer',
  'gpu%': 'gpu-recent',
  'Gpu%': 'gpu-longer',
  'gpumem%': 'gpumem-recent',
  'Gpumem%': 'gpumem-longer',
  'cpufail': 'cpu-status',
  'gpufail': 'gpu-status',
  'users': 'users-recent',
  'Users': 'users-longer',
  'jobs': 'jobs-recent',
  'Jobs': 'jobs-longer',
}

const builtinOperation: Record<string, any> = {}
builtinOperation['gpu'] = compileQuery('gpu*', knownFields, builtinOperation)
builtinOperation['compute'] = compileQuery('c*', knownFields, builtinOperation)
builtinOperation['hugemem'] = compileQuery('hugemem*', knownFields, builtinOperation)
builtinOperation['login'] = compileQuery('login*', knownFields, builtinOperation)
builtinOperation['cpu-busy'] = compileQuery('cpu% >= 50', knownFields, builtinOperation)
builtinOperation['Cpu-busy'] = compileQuery('Cpu% >= 50', knownFields, builtinOperation)
builtinOperation['cpu-idle'] = compileQuery('cpu% < 50', knownFields, builtinOperation)
builtinOperation['Cpu-idle'] = compileQuery('Cpu% < 50', knownFields, builtinOperation)
builtinOperation['virt-busy'] = compileQuery('virt% >= 50', knownFields, builtinOperation)
builtinOperation['Virt-busy'] = compileQuery('Virt% >= 50', knownFields, builtinOperation)
builtinOperation['virt-idle'] = compileQuery('virt% < 50', knownFields, builtinOperation)
builtinOperation['Virt-idle'] = compileQuery('Virt% < 50', knownFields, builtinOperation)
builtinOperation['res-busy'] = compileQuery('res% >= 50', knownFields, builtinOperation)
builtinOperation['Res-busy'] = compileQuery('Res% >= 50', knownFields, builtinOperation)
builtinOperation['res-idle'] = compileQuery('res% < 50', knownFields, builtinOperation)
builtinOperation['Res-idle'] = compileQuery('Res% < 50', knownFields, builtinOperation)
builtinOperation['gpu-busy'] = compileQuery('gpu% >= 50', knownFields, builtinOperation)
builtinOperation['Gpu-busy'] = compileQuery('Gpu% >= 50', knownFields, builtinOperation)
builtinOperation['gpu-idle'] = compileQuery('gpu and gpu% < 50', knownFields, builtinOperation)
builtinOperation['Gpu-idle'] = compileQuery('gpu and Gpu% < 50', knownFields, builtinOperation)
builtinOperation['gpumem-busy'] = compileQuery('gpumem% >= 50', knownFields, builtinOperation)
builtinOperation['Gpumem-busy'] = compileQuery('Gpumem% >= 50', knownFields, builtinOperation)
builtinOperation['gpumem-idle'] = compileQuery('gpu and gpumem% < 50', knownFields, builtinOperation)
builtinOperation['Gpumem-idle'] = compileQuery('gpu and Gpumem% < 50', knownFields, builtinOperation)
builtinOperation['cpu-down'] = compileQuery('cpufail > 0', knownFields, builtinOperation)
builtinOperation['gpu-down'] = compileQuery('gpu and gpufail > 0', knownFields, builtinOperation)
builtinOperation['busy'] =
  compileQuery('cpu-busy or gpu-busy or res-busy or virt-busy or gpumem-busy',
    knownFields, builtinOperation)
builtinOperation['Busy'] =
  compileQuery('Cpu-busy or Gpu-busy or Res-busy or Virt-busy or Gpumem-busy',
    knownFields, builtinOperation)
builtinOperation['idle'] =
  compileQuery('cpu-idle and virt-idle and res-idle and (~gpu* or (gpu% < 50 and gpumem% < 50))',
    knownFields, builtinOperation)
builtinOperation['Idle'] =
  compileQuery('Cpu-idle and Virt-idle and Res-idle and (~gpu* or (Gpu% < 50 and Gpumem% < 50))',
    knownFields, builtinOperation)
builtinOperation['down'] = compileQuery('cpu-down or gpu-down', knownFields, builtinOperation)

export function makeFilter(query: string) {
  const q: any = compileQuery(query, knownFields, builtinOperation)
  if (typeof q === 'string') {
    throw new Error(q)
  }
  return (d: any) => {
    const s = new Bitset(1)
    s.fill()
    const r = q.eval([d], s)
    return r.isSet(0)
  }
}

export const prepareJobQueryString = (jobQueryValues?: JobQueryValues) => {
  if(!jobQueryValues) {
    return ''
  }
  let query = `cluster=${jobQueryValues.clusterName}`

  const trimmedUsernames = jobQueryValues.usernames || ''

  const userNameList = trimmedUsernames.split(',').map(item => item.trim())
  if (userNameList?.length === 0) {
    query += '&user=-'
  } else {
    userNameList?.forEach(userName => {
      query += `&user=${userName}`
    })
  }

  const nodeNameList = jobQueryValues.nodeNames ? splitMultiPattern(jobQueryValues.nodeNames) : []
  nodeNameList.forEach(nodeName => {
    query += `&host=${nodeName}`
  })

  const jobIdList = jobQueryValues.jobIds ? jobQueryValues.jobIds.split(',').map(id => parseInt(id)) : []
  jobIdList.forEach(jobId => {
    query += `&job=${jobId}`
  })

  const fromDate = jobQueryValues.fromDate
  query += `&from=${fromDate}`

  const toDate = jobQueryValues.toDate
  query += `&to=${toDate}`

  const minRuntime = jobQueryValues.minRuntime
  if (minRuntime) {
    query += `&min-runtime=${minRuntime}`
  }

  const minPeakCpuCores = jobQueryValues.minPeakCpuCores
  if (minPeakCpuCores) {
    query += `&min-cpu-peak=${minPeakCpuCores * 100}`
  }

  const minPeakResidentGb = jobQueryValues.minPeakResidentGb
  if (minPeakResidentGb) {
    query += `&min-res-peak=${minPeakResidentGb}`
  }

  const gpuUsage = jobQueryValues.gpuUsage
  if (gpuUsage !== 'either') {
    query += `&${gpuUsage}=${TRUE_VAL}`
  }

  const fmt = 'fmt=json,job,user,host,duration,start,end,cpu-peak,res-peak,mem-peak,gpu-peak,gpumem-peak,cmd'
  query += `&${fmt}`

  return query
}

export const prepareShareableJobQueryLink = (jobQueryValues?: JobQueryValues) => {
  const queryString = prepareJobQueryString(jobQueryValues)
  const uri = `${window.location.origin}/jobQuery?${queryString}`
  return encodeURI(uri)
}

function isSpace(c: string) {
  return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

function isDelim(c: string) {
  return c == '(' || c == ')'
}

function isPunct(c: string) {
  return c == '<' || c == '>' || c == '=' || c == '~'
}

function isToken(c: string) {
  return !isDelim(c) && !isPunct(c) && !isSpace(c)
}

import { Bitset } from './Bitset.ts'

type Operator = '<' | '<=' | '>' | '>=' | '='

export class RelOperation {

  private readonly op: Operator
  private readonly fld: string
  private readonly n: number
  private readonly fn: (a: number) => boolean

  constructor(op: Operator, fld: any, n: number) {
    if (!(typeof fld === 'string' && typeof n === 'number')) {
      throw new Error(`Wrong type of arguments to relational operator ${op}`)
    }
    this.op = op
    this.fld = fld
    this.n = n
    switch (op) {
    case '<':
      this.fn = (a) => a < n
      break
    case '<=':
      this.fn = (a) => a <= n
      break
    case '>':
      this.fn = (a) => a > n
      break
    case '>=':
      this.fn = (a) => a >= n
      break
    case '=':
      this.fn = (a) => a == n
      break
    default:
      throw new Error(`Internal error`)
    }
  }

  toString(): string {
    return `(${this.op} ${this.fld} ${this.n})`
  }

  eval(data: { [key: string]: any }[], elems: Bitset): Bitset {
    let result = new Bitset(elems.length)
    let fn = this.fn
    let fld = this.fld
    elems.enumerate((n: number) => {
      if (fn(data[n][fld])) {
        result.setBit(n)
      }
    })
    return result
  }
}

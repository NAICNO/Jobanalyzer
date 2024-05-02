import { Bitset } from './Bitset.ts'

export class NotOperation {
  e: any

  constructor(e: any) {
    this.e = e
  }

  toString(): string {
    return `(~ ${this.e})`
  }


// not(e) is (universe \ e)
  eval(data: any[], elems: Bitset): Bitset {
    let result = new Bitset(data.length)
    result.fill()
    let xs = this.e.eval(data, elems)
    xs.enumerate((n: number) => {
      result.clearBit(n)
    })
    return result
  }
}



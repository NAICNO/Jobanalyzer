import { Bitset, bitsetIntersection, bitsetUnion } from './Bitset.ts'

export class SetOperation {
  private readonly op
  private readonly e1
  private readonly e2

  constructor(op: any, e1: any, e2: any) {
    this.op = op
    this.e1 = e1
    this.e2 = e2
  }

  toString() {
    return `(${this.op} ${this.e1} ${this.e2})`
  }

  eval(data: any, elems: any): Bitset {
    if (this.op === 'or') {
      let v1 = this.e1.eval(data, elems)
      let v2 = this.e2.eval(data, elems)
      return bitsetUnion(v1, v2)
    }

    let v1 = this.e1.eval(data, elems)
    let v2 = this.e2.eval(data, elems)
    return bitsetIntersection(v1, v2)
  }

}

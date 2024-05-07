import { Bitset } from './Bitset.ts'
import { HostGlobber } from './HostGlobber.ts'

export class GlobOperation {
  private readonly glob: string
  private readonly matcher: HostGlobber

  constructor(glob: string) {
    this.glob = glob
    this.matcher = new HostGlobber(glob, true)
  }

  toString(): string {
    return `(node ${this.glob})`
  }

  eval(data: { hostname: string }[], elems: Bitset): Bitset {
    const result = new Bitset(elems.length)
    const matcher = this.matcher
    elems.enumerate((n: number) => {
      if (matcher.match(data[n].hostname)) {
        result.setBit(n)
      }
    })
    return result
  }
}

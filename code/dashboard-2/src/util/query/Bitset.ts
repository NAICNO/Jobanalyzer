export class Bitset {
  readonly elts: number[];
  public length: number;

  constructor(size: number, allSet: boolean = false) {
    this.elts = new Array(Math.ceil(size / 24)).fill(0);
    this.length = size;
    if (allSet) {
      this.fill();
    }
  }

  fill(): void {
    const full = Math.floor(this.length / 24);
    this.elts.fill(0xFFFFFF, 0, full);
    if ((this.length % 24) != 0) {
      this.elts[this.elts.length - 1] = (1 << (this.length % 24)) - 1;
    }
  }

  enumerate(f: (index: number) => void): void {
    for (let k = 0; k < this.elts.length; k++) {
      let v = this.elts[k];
      if (v !== 0) {
        for (let j = 0; j < 24; j++) {
          if (v & (1 << j)) {
            f((k * 24) + j);
          }
        }
      }
    }
  }

  setBit(n: number): void {
    this.elts[Math.floor(n / 24)] |= 1 << (n % 24);
  }

  clearBit(n: number): void {
    this.elts[Math.floor(n / 24)] &= ~(1 << (n % 24));
  }

  isSet(n: number): boolean {
    return (this.elts[Math.floor(n / 24)] & (1 << (n % 24))) !== 0;
  }

  toString(): string {
    const xs: number[] = [];
    this.enumerate(k => xs.push(k));
    return xs.toString();
  }

  toArray(): number[] {
    const xs = new Array(this.length).fill(0);
    this.enumerate(k => xs[k] = 1);
    return xs;
  }
}

export function bitsetUnion(a: Bitset, b: Bitset): Bitset {
  const result = new Bitset(Math.max(a.length, b.length));
  for (let i = 0; i < a.length; i++) {
    result.elts[i] = a.elts[i];
  }
  for (let i = 0; i < b.length; i++) {
    result.elts[i] |= b.elts[i];
  }
  return result;
}

export function bitsetIntersection(a: Bitset, b: Bitset): Bitset {
  // The reason the length is the max of the two lengths is because it simplifies the client for
  // all sets to have the same length, notably bitset complement requires knowing the set of the
  // universe.
  const result = new Bitset(Math.max(a.length, b.length));
  for (let i = 0; i < result.elts.length; i++) {
    let v0 = a.elts.length > i ? a.elts[i] : 0;
    let v1 = b.elts.length > i ? b.elts[i] : 0;
    result.elts[i] = v0 & v1;
  }
  return result;
}

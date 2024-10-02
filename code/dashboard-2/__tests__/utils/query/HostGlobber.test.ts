import { HostGlobber, splitMultiPattern } from '../../../src/util/query/HostGlobber.ts'

describe('HostGlobber', () => {
  it('should be tested', () => {
    let g = new HostGlobber('a*[10-30].b[2,4]')
    assert.isTrue(g.match('a15.b2'))
    assert.isTrue(g.match('abc27.b4'))
    assert.isFalse(g.match('a31.b2'))

    g = new HostGlobber('a*[1-5]c.b[10-20]d')
    assert.isTrue(g.match('a3c.b12d'))
    assert.isFalse(g.match('a3c.b21d'))
    assert.isFalse(g.match('a3x.b12d'))

    g = new HostGlobber('c[1-3]-[2,4]')
    assert.isTrue(g.match('c1-2'))
    assert.isTrue(g.match('c2-2'))
    assert.isFalse(g.match('c2-3'))

    assertExcept(() => new HostGlobber('a[]'), 'Invalid number')
    assertExcept(() => new HostGlobber('a[1-123456789012]'), 'Number out of range')
    assertExcept(() => new HostGlobber('a[1000000-1009999]'), 'Expression too large')
    assertExcept(() => new HostGlobber('a[4-3]'), 'Invalid range')
    assertExcept(() => new HostGlobber('a[1-50000]'), 'Range too large')
    assertExcept(() => new HostGlobber('a[1-5x]'), 'Expected \',\'')
    assertExcept(() => new HostGlobber('a[1'), 'Expected \',\'')
    assertExcept(() => new HostGlobber('a,b'), '\',\' not allowed here')

  })

  it('should be tested', () => {
    assert.deepEqual(splitMultiPattern(''), [])
    assert.deepEqual(splitMultiPattern('a'), ['a'])
    assert.deepEqual(splitMultiPattern('a,b'), ['a', 'b'])
    assert.deepEqual(splitMultiPattern('yes.no,ml[1-3].hi,ml[1,2],zappa'),
      ['yes.no', 'ml[1-3].hi', 'ml[1,2]', 'zappa'])

    assertExcept(() => splitMultiPattern('yes[hi'), 'Missing end bracket')
    assertExcept(() => splitMultiPattern('yes[hi[]'), 'nested brackets')
    assertExcept(() => splitMultiPattern('yes]'), 'unmatched end bracket')
    assertExcept(() => splitMultiPattern(',yes'), 'Empty host name')
    assertExcept(() => splitMultiPattern('yes,'), 'Empty host name')
    assertExcept(() => splitMultiPattern('yes,,no'), 'Empty host name')
  })
})

function assertExcept(thunk: () => void, s: string): void {
  try {
    thunk()
    throw new Error(`Expected exception with payload "${s}"`)
  } catch (e: any) {
    if (String(e).indexOf(s) === -1) {
      throw new Error(`Expected payload "${s}" got "${String(e)}"`)
    }
  }
}


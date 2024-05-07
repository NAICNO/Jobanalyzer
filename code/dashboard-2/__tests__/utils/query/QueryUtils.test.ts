import { Bitset } from '../../../src/util/query/Bitset.ts'
import { compileQuery } from '../../../src/util/query/QueryUtils.ts'

describe('QueryUtils', () => {
  it('should test Queryterm', () => {
    let data = [{x: 5}, {x: 10}, {x: 20}, {x: 30}, {x: 7}]
    let elems = new Bitset(data.length, true)

    let q_lt = compileQuery('x < 20', {'x': true})
    assert.deepEqual(q_lt.eval(data, elems).toArray(), [1, 1, 0, 0, 1])

    let q_le = compileQuery('x <= 20', {'x': true})
    assert.deepEqual(q_le.eval(data, elems).toArray(), [1, 1, 1, 0, 1])

    let q_gt = compileQuery('x > 20', {'x': true})
    assert.deepEqual(q_gt.eval(data, elems).toArray(), [0, 0, 0, 1, 0])

    let q_ge = compileQuery('x >= 20', {'x': true})
    assert.deepEqual(q_ge.eval(data, elems).toArray(), [0, 0, 1, 1, 0])

    let q_eq = compileQuery('x = 20', {'x': true})
    assert.deepEqual(q_eq.eval(data, elems).toArray(), [0, 0, 1, 0, 0])

    let q_not = compileQuery('~(x = 20)', {'x': true})
    assert.deepEqual(q_not.eval(data, elems).toArray(), [1, 1, 0, 1, 1])

    let r = compileQuery('x = 10')
    assert.isTrue(typeof r == 'string' && r.indexOf('Wrong type') != -1)

    r = compileQuery('x = y', {'x': true, 'y': true})
    assert.isTrue(typeof r == 'string' && r.indexOf('Wrong type') != -1)
  })

  it('should test Host globs', () => {

    // We assume matching itself is implemented correctly: HostGlobber.ts.

    let data = [{hostname: 'c1-1'}, {hostname: 'c2-1'}, {hostname: 'c1-37'}]
    let elems = new Bitset(data.length, true)

    let q_glob = compileQuery('c1-*')
    assert.deepEqual(q_glob.eval(data, elems).toArray(), [1, 0, 1])

    let q_notglob = compileQuery('~c1-*')
    assert.deepEqual(q_notglob.eval(data, elems).toArray(), [0, 1, 0])
  })

  it('should test Conjunction/disjunction, precedence of binops over and/or', () => {
    let data = [{x: 5, y: 10}, {x: 10, y: 1}, {x: 20, y: 9}, {x: 30, y: 1}, {x: 7, y: 5}]
    let elems = new Bitset(data.length, true)

    let q_and = compileQuery('x < 20 and y > 5', {'x': true, 'y': true})
    assert.deepEqual(q_and.eval(data, elems).toArray(), [1, 0, 0, 0, 0])

    let q_or = compileQuery('x < 20 or y > 5', {'x': true, 'y': true})
    assert.deepEqual(q_or.eval(data, elems).toArray(), [1, 1, 1, 0, 1])
  })

  it(' should test relative precedence of and/or', () => {
    // These are constructed so that the binding of and/or will give different results
    let data = [{x: 1, y: 2, z: 3}, {x: 1, y: 3, z: 4}, {x: 2, y: 3, z: 4}]
    let elems = new Bitset(data.length, true)

    let q_or = compileQuery('x = 1 and (y = 2 or z = 4)', {'x': true, 'y': true, 'z': true})
    assert.deepEqual(q_or.eval(data, elems).toArray(), [1, 1, 0])

    let q_and = compileQuery('x = 1 and y = 2 or z = 4', {'x': true, 'y': true, 'z': true})
    assert.deepEqual(q_and.eval(data, elems).toArray(), [1, 1, 1])
  })

})

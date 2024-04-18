import '@testing-library/jest-dom'
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { HostNameFieldCell } from '../../../src/components/table/HostNameFieldCell'

describe('HostNameFieldCell', () => {
  it('renders correctly when value is in HOSTNAMES_ALIAS', () => {
    const value = 'ml1.hpc.uio.no' // assuming 1 is a key in HOSTNAMES_ALIAS
    const {asFragment} = render(<HostNameFieldCell value={value}/>)
    expect(asFragment()).toMatchSnapshot()
    expect(screen.getByText('ML1').parentNode)
  })

  it('renders correctly when value is not in HOSTNAMES_ALIAS', () => {
    const value = '999' // assuming 999 is not a key in HOSTNAMES_ALIAS
    const {asFragment} = render(<HostNameFieldCell value={value}/>)
    expect(asFragment()).toMatchSnapshot()
  })
})

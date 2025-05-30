import { render as rtlRender } from '@testing-library/react'
import { Provider } from '../src/components/ui/provider.tsx'
import React from 'react'

export function render(ui: React.ReactNode) {
  return rtlRender(<>{ui}</>, {
    wrapper: (props: React.PropsWithChildren) => (
      <Provider>{props.children}</Provider>
    ),
  })
}

import { ReactElement } from 'react'
import { ChakraProvider, extendTheme } from '@chakra-ui/react'
import { render } from '@testing-library/react'

export const chakraRender = (ui: ReactElement, {colorMode = 'light'}) => {
  return render(
    <ChakraProvider theme={extendTheme({config: {initialColorMode: colorMode}})}>
      {ui}
    </ChakraProvider>
  )
}

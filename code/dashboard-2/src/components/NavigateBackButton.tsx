import { Link, useNavigate, useLocation } from 'react-router'
import { IconButton } from '@chakra-ui/react'
import { MdArrowBack } from 'react-icons/md'

export const NavigateBackButton = () => {

  const navigate = useNavigate()
  const location = useLocation()

  const goBack = () => {
    if (location.key === 'default') {
      navigate('/')
    } else {
      navigate(-1)
    }
  }

  return (
    <IconButton
      variant={'outline'}
      colorPalette={'blue'}
      rounded="full"
      size={{base: 'sm', md: 'md'}}
      aria-label="Back"
      as={Link}
      onClick={goBack}
    >
      <MdArrowBack/>
    </IconButton>
  )
}

import { ArrowBackIcon } from '@chakra-ui/icons'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { IconButton } from '@chakra-ui/react'

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
      isRound={true}
      icon={<ArrowBackIcon boxSize={{base: 4, md: 6}}/>}
      aria-label="Back"
      as={Link}
      onClick={goBack}
    />
  )
}

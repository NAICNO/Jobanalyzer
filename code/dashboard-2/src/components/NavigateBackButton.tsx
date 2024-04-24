import { ArrowBackIcon } from '@chakra-ui/icons'
import { Link, useNavigate } from 'react-router-dom'
import { IconButton } from '@chakra-ui/react'

export const NavigateBackButton = () => {

  const navigate = useNavigate()

  return (
    <IconButton
      isRound={true}
      icon={<ArrowBackIcon boxSize={{base: 4, md: 6}}/>}
      aria-label="Back"
      as={Link}
      onClick={() => navigate(-1)}
    />
  )
}

import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { APP_NAME, PAGE_TITLE_SUFFIX } from '../Constants.ts'

interface PageTitleProps {
  title: string
}

export const PageTitle = ({title}: PageTitleProps) => {
  const location = useLocation()

  const pageTitle = title ? `${title}${PAGE_TITLE_SUFFIX}` : APP_NAME

  useEffect(() => {
    document.title = pageTitle
  }, [location, title])

  return null
}


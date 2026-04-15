export const APP_NAME = 'NAIC Jobanalyzer'

export const PAGE_TITLE_SUFFIX = ' | ' + APP_NAME

const _base = import.meta.env.BASE_URL || '/'
export const APP_BASE_PREFIX = _base.endsWith('/') ? _base : _base + '/'

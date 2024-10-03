import * as yup from 'yup'
import { JOB_QUERY_VALIDATION_SCHEMA } from '../Constants.ts'

export interface JobQueryValues extends yup.InferType<typeof JOB_QUERY_VALIDATION_SCHEMA> {
  // using InferType to get extract the interface from the schema
}

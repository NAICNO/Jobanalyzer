import { useEffect, useState } from 'react'
import {
  Box,
  Button,
  Checkbox,
  CheckboxGroup,
  Field as ChakraField,
  Dialog,
  RadioGroup,
  VStack,
  DialogOpenChangeDetails,
} from '@chakra-ui/react'
import { Field, FieldProps, Form, Formik } from 'formik'

import {
  JOB_QUERY_EXPORT_FORMATS,
  JOB_QUERY_EXPORT_VALIDATION_SCHEMA,
  JOB_QUERY_RESULTS_COLUMN,
} from '../Constants.ts'
import { useExportJobQuery } from '../hooks/useExportJobQuery.ts'
import { ExportOptions, JobQueryValues } from '../types'

const downloadExported = (exportContent: string, filename: string, mimeType: string) => {
  const file = new Blob([exportContent], {type: mimeType})
  const url = URL.createObjectURL(file)

  const element = document.createElement('a')

  element.href = url
  element.download = filename
  document.body.appendChild(element)

  element.click()

  setTimeout(() => {
    document.body.removeChild(element)
    URL.revokeObjectURL(url)
  }, 200)
}

const columnList = Object.values(JOB_QUERY_RESULTS_COLUMN)

interface JobQueryResultExportModalProps {
  open: boolean
  onClose: () => void
  jobQueryFormValues: JobQueryValues
}

const initialExportOptions: ExportOptions = {
  format: JOB_QUERY_EXPORT_FORMATS[0].value,
  fields: [],
}

const JobQueryResultExportModal = ({open, onClose, jobQueryFormValues}: JobQueryResultExportModalProps) => {

  const [exportOptions, setExportOptions] = useState(initialExportOptions)

  const {refetch, data, isFetching} = useExportJobQuery(jobQueryFormValues, exportOptions)

  useEffect(() => {
    if (data) {
      const exportFormat = JOB_QUERY_EXPORT_FORMATS.find((format) => format.value === exportOptions.format)
      downloadExported(data, `download.${exportFormat?.fileExtension}`, exportFormat?.mimeType)
      onClose()
    }
  }, [data])

  const handleOpenChange = (details: DialogOpenChangeDetails) => {
    if (!details.open) {
      onClose()
    }
  }


  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <Dialog.Backdrop/>
      <Formik
        enableReinitialize={true}
        initialValues={initialExportOptions}
        validationSchema={JOB_QUERY_EXPORT_VALIDATION_SCHEMA}
        onSubmit={(exportOptions) => {
          setExportOptions(exportOptions)
          setTimeout(() => {
            refetch()
          }, 0)
        }}>
        {({isValid, submitForm}) => (
          <Dialog.Content>
            <Dialog.Header>Export Result</Dialog.Header>
            <Dialog.CloseTrigger/>
            <Dialog.Body>
              <Form>
                <Field name="format">
                  {({field, meta, form}: FieldProps) => (
                    <ChakraField.Root as="fieldset" mb="20px">
                      <ChakraField.Label as="legend">File format</ChakraField.Label>
                      <RadioGroup.Root
                        {...field}
                        value={field.value}
                        onChange={val => form.setFieldValue('format', val)}
                      >
                        <VStack alignItems={'start'}>
                          {
                            JOB_QUERY_EXPORT_FORMATS.map((format) => (
                              <RadioGroup.Item key={format.value} value={format.value}>
                                {format.label}
                              </RadioGroup.Item>
                            ))
                          }
                        </VStack>
                      </RadioGroup.Root>
                      <ChakraField.ErrorText>{meta.error}</ChakraField.ErrorText>
                    </ChakraField.Root>
                  )}
                </Field>

                <Field name="fields">
                  {({field, meta, form}: FieldProps) => (
                    <ChakraField.Root as="fieldset" invalid={!!(meta.error && meta.touched)}>
                      <ChakraField.Label as="legend">Fields</ChakraField.Label>
                      <Checkbox.Root
                        mb="10px"
                        checked={field.value.length > 0 && field.value.length < columnList.length ? 'indeterminate' : true}
                        onCheckedChange={(e) => {
                          if (e.checked) {
                            form.setFieldValue('fields', columnList.map((column) => column.key))
                          } else {
                            form.setFieldValue('fields', [])
                          }
                        }}
                      >
                        <Checkbox.HiddenInput/>
                        <Checkbox.Control>
                          <Checkbox.Indicator/>
                        </Checkbox.Control>
                        <Checkbox.Label>Select all</Checkbox.Label>
                      </Checkbox.Root>
                      <Box ml="20px">
                        <CheckboxGroup
                          {...field}
                          value={field.value}
                          onChange={val => form.setFieldValue('fields', val)}
                        >
                          <VStack alignItems={'start'}>
                            {
                              columnList.map((column) => (
                                <Checkbox.Root key={column.key} value={column.key}>
                                  {column.title}
                                </Checkbox.Root>
                              ))
                            }
                          </VStack>
                        </CheckboxGroup>
                      </Box>
                      <ChakraField.ErrorText>{meta.error}</ChakraField.ErrorText>
                    </ChakraField.Root>
                  )}
                </Field>
              </Form>
            </Dialog.Body>
            <Dialog.Footer>
              <Button
                colorScheme="blue"
                mr={3}
                onClick={submitForm}
                disabled={!isValid || isFetching}
                loading={isFetching}
              >
                Download
              </Button>
              <Button
                variant="ghost"
                onClick={onClose}
              >
                Close
              </Button>
            </Dialog.Footer>
          </Dialog.Content>
        )}
      </Formik>
    </Dialog.Root>
  )
}
export default JobQueryResultExportModal

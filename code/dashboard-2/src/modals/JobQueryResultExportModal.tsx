import {
  Box,
  Button, Checkbox, CheckboxGroup, FormControl, FormErrorMessage, FormLabel,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay, Radio, RadioGroup,
  VStack,
} from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import JobQueryValues from '../types/JobQueryValues.ts'
import {
  JOB_QUERY_EXPORT_FORMATS,
  JOB_QUERY_EXPORT_VALIDATION_SCHEMA,
  JOB_QUERY_RESULTS_COLUMN,
} from '../Constants.ts'
import { useExportJobQuery } from '../hooks/useExportJobQuery.ts'
import { Field, FieldProps, Form, Formik } from 'formik'

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
  isOpen: boolean
  onClose: () => void
  jobQueryFormValues: JobQueryValues
}

const initialExportOptions: ExportOptions = {
  format: JOB_QUERY_EXPORT_FORMATS[0].value,
  fields: [],
}

const JobQueryResultExportModal = ({isOpen, onClose, jobQueryFormValues}: JobQueryResultExportModalProps) => {

  const [exportOptions, setExportOptions] = useState(initialExportOptions)

  const {refetch, data, isFetching} = useExportJobQuery(jobQueryFormValues, exportOptions)

  useEffect(() => {
    if (data) {
      const exportFormat = JOB_QUERY_EXPORT_FORMATS.find((format) => format.value === exportOptions.format)
      downloadExported(data, `download.${exportFormat?.fileExtension}`, exportFormat?.mimeType)
      onClose()
    }
  }, [data])


  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay/>
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
          <ModalContent>
            <ModalHeader>Export Result</ModalHeader>
            <ModalCloseButton/>
            <ModalBody>
              <Form>
                <Field name="format">
                  {({field, meta, form}: FieldProps) => (
                    <FormControl as="fieldset" mb="20px">
                      <FormLabel as="legend">File format</FormLabel>
                      <RadioGroup
                        {...field}
                        value={field.value}
                        onChange={val => form.setFieldValue('format', val)}
                      >
                        <VStack alignItems={'start'}>
                          {
                            JOB_QUERY_EXPORT_FORMATS.map((format) => (
                              <Radio key={format.value} value={format.value}>
                                {format.label}
                              </Radio>
                            ))
                          }
                        </VStack>
                      </RadioGroup>
                      <FormErrorMessage>{meta.error}</FormErrorMessage>
                    </FormControl>
                  )}
                </Field>

                <Field name="fields">
                  {({field, meta, form}: FieldProps) => (
                    <FormControl as="fieldset" isInvalid={!!(meta.error && meta.touched)}>
                      <FormLabel as="legend">Fields</FormLabel>
                      <Checkbox
                        mb="10px"
                        isIndeterminate={field.value.length > 0 && field.value.length < columnList.length}
                        onChange={(e) => {
                          if (e.target.checked) {
                            form.setFieldValue('fields', columnList.map((column) => column.key))
                          } else {
                            form.setFieldValue('fields', [])
                          }
                        }}
                      >
                        Select all
                      </Checkbox>
                      <Box ml="20px">
                        <CheckboxGroup
                          {...field}
                          value={field.value}
                          onChange={val => form.setFieldValue('fields', val)}
                        >
                          <VStack alignItems={'start'}>
                            {
                              columnList.map((column) => (
                                <Checkbox key={column.key} value={column.key}>
                                  {column.title}
                                </Checkbox>
                              ))
                            }
                          </VStack>
                        </CheckboxGroup>
                      </Box>
                      <FormErrorMessage>{meta.error}</FormErrorMessage>
                    </FormControl>
                  )}
                </Field>
              </Form>
            </ModalBody>
            <ModalFooter>
              <Button
                colorScheme="blue"
                mr={3}
                onClick={submitForm}
                isDisabled={!isValid || isFetching}
                isLoading={isFetching}
              >
                Download
              </Button>
              <Button
                variant="ghost"
                onClick={onClose}
              >
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        )}
      </Formik>
    </Modal>
  )
}
export default JobQueryResultExportModal

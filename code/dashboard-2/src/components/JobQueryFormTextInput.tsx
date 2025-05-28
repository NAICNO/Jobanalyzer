import { Field, FieldProps } from 'formik'
import { Field as ChakraField, Input } from '@chakra-ui/react'

interface JobQueryFormTextInputProps {
  name: string;
  label: string;
  type: string;
  placeholder: string;
}

export const JobQueryFormTextInput = ({name, label, type, placeholder}: JobQueryFormTextInputProps) => {
  return (
    <Field name={name}>
      {({field, meta}: FieldProps) => (
        <ChakraField.Root invalid={!!(meta.error && meta.touched)}>
          <ChakraField.Label>{label}</ChakraField.Label>
          <Input {...field} type={type} placeholder={placeholder} value={field.value} data-bwignore/>
          <ChakraField.ErrorText>{meta.error}</ChakraField.ErrorText>
        </ChakraField.Root>
      )}
    </Field>
  )
}

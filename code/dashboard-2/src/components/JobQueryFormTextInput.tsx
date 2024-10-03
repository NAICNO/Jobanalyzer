import { Field, FieldProps } from 'formik'
import { FormControl, FormErrorMessage, FormLabel, Input } from '@chakra-ui/react'

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
        <FormControl isInvalid={!!(meta.error && meta.touched)}>
          <FormLabel>{label}</FormLabel>
          <Input {...field} type={type} placeholder={placeholder} value={field.value} data-bwignore/>
          <FormErrorMessage>{meta.error}</FormErrorMessage>
        </FormControl>
      )}
    </Field>
  )
}

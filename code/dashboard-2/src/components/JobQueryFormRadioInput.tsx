import { Field, FieldProps } from 'formik'
import { FormControl, FormLabel, HStack, Radio, RadioGroup } from '@chakra-ui/react'
import { SimpleRadioInputOption } from '../types'

interface JobQueryFormRadioInputProps {
  name: string;
  label: string;
  options: SimpleRadioInputOption[];
}

export const JobQueryFormRadioInput = ({name, label, options}: JobQueryFormRadioInputProps) => {
  return (
    <Field name={name}>
      {({field, form}: FieldProps) => (
        <FormControl as="fieldset">
          <FormLabel as="legend">{label}</FormLabel>
          <RadioGroup
            {...field}
            value={field.value}
            onChange={val => form.setFieldValue(name, val)}
          >
            <HStack spacing="24px">
              {
                options.map((option) => (
                  <Radio key={option.value} value={option.value}>
                    {option.text}
                  </Radio>
                ))
              }
            </HStack>
          </RadioGroup>
        </FormControl>
      )}
    </Field>
  )
}

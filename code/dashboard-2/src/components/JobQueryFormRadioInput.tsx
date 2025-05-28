import { Field, FieldProps } from 'formik'
import { Field as ChakraField, HStack, RadioGroup } from '@chakra-ui/react'
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
        <ChakraField.Root as="fieldset">
          <ChakraField.Label as="legend" pb={2}>{label}</ChakraField.Label>
          <RadioGroup.Root
            {...field}
            value={field.value}
            onValueChange={val => form.setFieldValue(name, val)}
          >
            <HStack gap="24px">
              {
                options.map((option) => (
                  <RadioGroup.Item key={option.value} value={option.value}>
                    <RadioGroup.ItemHiddenInput/>
                    <RadioGroup.ItemIndicator/>
                    <RadioGroup.ItemText>{option.text}</RadioGroup.ItemText>
                  </RadioGroup.Item>
                ))
              }
            </HStack>
          </RadioGroup.Root>
        </ChakraField.Root>
      )}
    </Field>
  )
}

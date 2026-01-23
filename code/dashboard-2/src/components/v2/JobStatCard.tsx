import { memo } from 'react'
import { Card, Stat, HStack, Text, Tooltip, Icon } from '@chakra-ui/react'
import { HiInformationCircle } from 'react-icons/hi2'

interface JobStatCardProps {
  label: string
  value: string | number
  helpText?: string
  tooltip?: string
}

export const JobStatCard = memo(({ label, value, helpText, tooltip }: JobStatCardProps) => {
  return (
    <Card.Root size="sm">
      <Card.Body>
        <Stat.Root>
          <Stat.Label fontSize="sm" color="fg.muted">
            <HStack gap={1}>
              <Text>{label}</Text>
              {tooltip && (
                <Tooltip.Root openDelay={300}>
                  <Tooltip.Trigger asChild>
                    <Icon size="md" color="gray.400" cursor="help">
                      <HiInformationCircle />
                    </Icon>
                  </Tooltip.Trigger>
                  <Tooltip.Positioner>
                    <Tooltip.Content maxW="300px">
                      <Text fontSize="xs">{tooltip}</Text>
                    </Tooltip.Content>
                  </Tooltip.Positioner>
                </Tooltip.Root>
              )}
            </HStack>
          </Stat.Label>
          <Stat.ValueText fontSize="2xl" fontWeight="bold">
            {value}
          </Stat.ValueText>
          {helpText && (
            <Stat.HelpText fontSize="xs" color="fg.muted">
              {helpText}
            </Stat.HelpText>
          )}
        </Stat.Root>
      </Card.Body>
    </Card.Root>
  )
})

JobStatCard.displayName = 'JobStatCard'

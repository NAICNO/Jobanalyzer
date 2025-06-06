import {
  Heading,
  Link as ChakraLink,
  List,
  Table,
  Text,
  VStack,
} from '@chakra-ui/react'
import { LuExternalLink } from 'react-icons/lu'

const fieldNames = [
  {name: 'cpu%', desc: 'cpu-recent'},
  {name: 'Cpu%', desc: 'cpu-longer (etc)'},
  {name: 'virt%', desc: 'virt-recent'},
  {name: 'res%', desc: 'res-recent'},
  {name: 'gpu%', desc: 'gpu-recent'},
  {name: 'gpumem%', desc: 'gpumem-recent field (this is physical RAM)'},
  {name: 'cpufail', desc: 'cpu-status'},
  {name: 'gpufail', desc: 'gpu-status'},
  {name: 'users', desc: 'users-recent'},
  {name: 'jobs', desc: 'jobs-recent'},
  {name: 'violators', desc: 'violators'},
  {name: 'zombies', desc: 'zombies'},
]

const abbreviations = [
  {name: 'compute', desc: 'c*'},
  {name: 'gpu', desc: 'gpu*'},
  {name: 'hugemem', desc: 'hugemem*'},
  {name: 'login', desc: 'login*'},
  {name: 'cpu-busy', desc: 'cpu% >= 50'},
  {name: 'cpu-idle', desc: 'cpu% < 50'},
  {name: 'virt-busy', desc: 'virt% >= 50'},
  {name: 'virt-idle', desc: 'virt% < 50'},
  {name: 'res-busy', desc: 'res% >= 50'},
  {name: 'res-idle', desc: 'res% < 50'},
  {name: 'gpu-busy', desc: 'gpu and gpu% >= 50'},
  {name: 'gpu-idle', desc: 'gpu and gpu% < 50'},
  {name: 'gpumem-busy', desc: 'gpu and gpumem% >= 50'},
  {name: 'gpumem-idle', desc: 'gpu and gpumem% < 50'},
  {name: 'cpu-down', desc: 'cpufail > 0'},
  {name: 'gpu-down', desc: 'gpu and gpufail > 0'},
  {name: 'busy', desc: 'cpu-busy or gpu-busy or virt-busy or res-busy or gpumem-busy'},
  {name: 'idle', desc: 'cpu-idle and virt-idle and res-idle and (~gpu or gpu-idle and gpumem-idle)'},
  {name: 'down', desc: 'cpu-down or gpu-down'},
]

export const NodeSelectionHelpContent = () => {
  return (
    <VStack alignItems="start" gap={2}>
      <Text>The query expression selects a subset of all nodes by applying filters.</Text>
      <Heading as="h4" size="md" mt="20px">Expressions</Heading>
      <Text>Query expression syntax is pretty simple. These are all expressions:</Text>
      <List.Root gap={2}>
        <List.Item>
          <Text>
            A <Text as="em">hostname glob</Text> is a wildcard expression selecting some hosts where "*" is a
            wildcard that stands for any number of characters, ie, "c1-*" selects all nodes in the
            "c1" group of nodes, while "c1*" selects the "c1", "c10", and "c11" groups. "*" by itself
            selects all nodes.
          </Text>
        </List.Item>
        <List.Item>
          <Text>
            An <Text as="em">abbreviation</Text> is a word that stands for a pre-defined expression, see the
            list below. For example, the abbreviation "busy" stands for a complex expression that
            selects all nodes that are deemed busy.
          </Text>
        </List.Item>
        <List.Item>
          <Text>
            A <Text as="em">relational expression</Text> on the form <Text as="code">fieldname <Text
            as="b">relation</Text> value </Text>
            selects nodes whose field <Text as="em">fieldname</Text> has a numeric <Text as="em">value</Text> that
            satisfies
            the relational operator, for example, "cpu% {'>'} 50" means that the node must be using more
            than 50% of its CPU capacity. The relational operators are "{'<'}", "{'<='}", "{'>'}", "{'>'}=",
            and "=". The field names are listed below.
          </Text>
        </List.Item>
        <List.Item>
          <Text>
            The <Text as="em">logical operations</Text> <Text as="b">and</Text> and <Text as="b">or</Text> are
            used to combine query expressions, and
            parentheses <Text as="b">(</Text> and <Text as="b">)</Text> are used to group them:
            <Text as="code">login* and (cpu% {'>'} 50 or mem% {'>'} 50)</Text>.
          </Text>
        </List.Item>
        <List.Item>
          <Text>
            A set of selected nodes can be complemented by the <Text as="b">~</Text> operator,
            eg, <Text as="code">~login*</Text> is any node except the login nodes.
          </Text>
        </List.Item>
      </List.Root>
      <Heading as="h4" size="md" mt="20px">Field names</Heading>
      <Text>
        The field names currently defined for the dashboard are those that appear in the table on the
        dashboard. The "recent" columns have uncapitalized names ("cpu%") while the "longer" columns have
        capitalized names ("Cpu%").
      </Text>
      <Table.ScrollArea mt="10px">
        <Table.Root size="sm">
          <Table.Header bg="gray.100">
            <Table.Row>
              <Table.ColumnHeader>Field</Table.ColumnHeader>
              <Table.ColumnHeader>Description</Table.ColumnHeader>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {
              fieldNames.map((field) => (
                <Table.Row key={field.name}>
                  <Table.Cell>{field.name}</Table.Cell>
                  <Table.Cell>{field.desc}</Table.Cell>
                </Table.Row>
              ))
            }
          </Table.Body>
        </Table.Root>
      </Table.ScrollArea>

      <Heading as="h4" size="md" mt="20px">Abbreviations</Heading>
      <Text>The predefined abbreviations are these:</Text>
      <Table.ScrollArea my="20px">
        <Table.Root size="sm">
          <Table.Header bg="gray.100">
            <Table.Row>
              <Table.ColumnHeader>Abbreviation</Table.ColumnHeader>
              <Table.ColumnHeader>Description</Table.ColumnHeader>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {
              abbreviations.map((field) => (
                <Table.Row key={field.name}>
                  <Table.Cell>{field.name}</Table.Cell>
                  <Table.Cell>{field.desc}</Table.Cell>
                </Table.Row>
              ))
            }
          </Table.Body>
        </Table.Root>
      </Table.ScrollArea>

      <Text>These have Capitalized variants for the "*-longer" data where that makes sense, eg, "Idle")</Text>
      <Text> For example, to find nodes with spare capacity, simply run "idle". ("Idle" is a poor
        moniker for something running at 49% utilization, so perhaps we want something more subtle.{' '}
        <ChakraLink href="https://github.com/NAICNO/Jobanalyzer/issues/new" target="_blank">
          File an issue here.
          <LuExternalLink/>
        </ChakraLink>
      </Text>

      <Text> Perhaps you want hugemem nodes with regular compute capacity: try "hugemem and idle".</Text>

      <Text> It's easy to add abbreviations - but at this time the abbreviations must be added in the
        query engine, they can't be added by the user.{' '}
        <ChakraLink href="https://github.com/NAICNO/Jobanalyzer/issues/new" target="_blank">
          File an issue here.
          <LuExternalLink/>
        </ChakraLink>
      </Text>
    </VStack>
  )
}

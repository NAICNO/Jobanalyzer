import { useCallback, useEffect, useMemo, useRef, useState, memo } from 'react'
import {
  Background,
  BackgroundVariant,
  Controls,
  Handle,
  MiniMap,
  Panel,
  Position,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
} from '@xyflow/react'
import type { Node, NodeProps } from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { Box, VStack, HStack, Text, Spinner, Alert, Badge, IconButton, SegmentGroup, Slider } from '@chakra-ui/react'
import { MdOutlineAlignHorizontalLeft, MdOutlineAlignVerticalTop } from 'react-icons/md'
import { LuPlay, LuPause } from 'react-icons/lu'
import dayjs from 'dayjs'
import type { Client } from '../../client/client/types.gen'
import { useColorMode } from '../ui/color-mode'
import { useJobProcessTree } from '../../hooks/v2/useJobQueries'
import {
  processTreeToFlow,
  processTreeToFlowAtTime,
  collectUniqueSampleTimes,
  getLayoutedElements,
  type ProcessNodeData,
} from '../../util/processTreeTransformers'

interface ProcessTreeViewProps {
  cluster: string
  jobId: number
  client: Client | null
  showTimeline?: boolean
}

const formatKiB = (kib: number): string => {
  if (kib >= 1048576) return `${(kib / 1048576).toFixed(1)} GiB`
  if (kib >= 1024) return `${(kib / 1024).toFixed(1)} MiB`
  return `${kib.toFixed(0)} KiB`
}

const ProcessNode = memo(({ data, targetPosition, sourcePosition }: NodeProps<Node<ProcessNodeData>>) => {
  return (
    <>
      <Handle type="target" position={targetPosition ?? Position.Top} />
      <VStack gap={0} align="start" p={2} minW="180px">
        <Text fontSize="xs" fontWeight="bold" truncate maxW="170px">
          {data.cmd}
        </Text>
        <Text fontSize="xs" color="fg.muted">
          PID: {data.pid}
        </Text>
        <HStack gap={2} mt={1}>
          {data.cpuUtil !== null && (
            <Badge size="xs" colorPalette="blue" variant="subtle">
              CPU {data.cpuUtil.toFixed(1)}%
            </Badge>
          )}
          {data.memoryResident !== null && (
            <Badge size="xs" colorPalette="green" variant="subtle">
              {formatKiB(data.memoryResident)}
            </Badge>
          )}
        </HStack>
      </VStack>
      <Handle type="source" position={sourcePosition ?? Position.Bottom} />
    </>
  )
})

ProcessNode.displayName = 'ProcessNode'

const nodeTypes = { processNode: ProcessNode }

interface TimelineSliderProps {
  steps: number[]
  currentStepIndex: number
  onStepChange: (index: number) => void
  isPlaying: boolean
  onPlayToggle: () => void
  processCount: number
  totalProcesses: number
}

const TimelineSlider = memo(({
  steps,
  currentStepIndex,
  onStepChange,
  isPlaying,
  onPlayToggle,
  processCount,
  totalProcesses,
}: TimelineSliderProps) => {
  const currentTime = steps[currentStepIndex] ?? steps[0]
  const formattedTime = dayjs.unix(currentTime).format('HH:mm:ss')
  const formattedStart = dayjs.unix(steps[0]).format('HH:mm:ss')
  const formattedEnd = dayjs.unix(steps[steps.length - 1]).format('HH:mm:ss')

  const handleSliderChange = useCallback((details: { value: number[] }) => {
    onStepChange(details.value[0])
  }, [onStepChange])

  const marks = useMemo(() => {
    const maxMarks = 20
    const total = steps.length
    if (total <= maxMarks) {
      return Array.from({ length: total }, (_, i) => i)
    }
    const interval = Math.ceil(total / maxMarks)
    const result: number[] = []
    for (let i = 0; i < total; i += interval) {
      result.push(i)
    }
    if (result[result.length - 1] !== total - 1) {
      result.push(total - 1)
    }
    return result
  }, [steps.length])

  return (
    <HStack
      w="100%"
      px={4}
      py={2}
      gap={3}
      borderTopWidth="1px"
      borderColor="border"
      bg="bg"
    >
      <IconButton
        aria-label={isPlaying ? 'Pause' : 'Play'}
        size="xs"
        variant="ghost"
        onClick={onPlayToggle}
      >
        {isPlaying ? <LuPause /> : <LuPlay />}
      </IconButton>

      <Text fontSize="xs" color="fg.muted" whiteSpace="nowrap">
        {formattedStart}
      </Text>

      <Slider.Root
        flex="1"
        size="sm"
        min={0}
        max={steps.length - 1}
        step={1}
        value={[currentStepIndex]}
        onValueChange={handleSliderChange}
      >
        <Slider.Control>
          <Slider.Track>
            <Slider.Range />
          </Slider.Track>
          <Slider.Thumb index={0}>
            <Slider.HiddenInput />
            <Slider.DraggingIndicator
              layerStyle="fill.solid"
              top="-8"
              rounded="sm"
              px="1.5"
              fontSize="xs"
            >
              {formattedTime}
            </Slider.DraggingIndicator>
          </Slider.Thumb>
          <Slider.Marks marks={marks} />
        </Slider.Control>
      </Slider.Root>

      <Text fontSize="xs" color="fg.muted" whiteSpace="nowrap">
        {formattedEnd}
      </Text>

      <Badge size="xs" variant="subtle" colorPalette="blue">
        {formattedTime}
      </Badge>
      <Badge size="xs" variant="subtle" colorPalette="green">
        {processCount}/{totalProcesses}
      </Badge>
    </HStack>
  )
})

TimelineSlider.displayName = 'TimelineSlider'

const ProcessTreeFlowInner = memo(({ cluster, jobId, client, showTimeline = false }: ProcessTreeViewProps) => {
  const [direction, setDirection] = useState<'TB' | 'LR'>('TB')
  const { colorMode } = useColorMode()
  const { fitView } = useReactFlow()

  const { data: treeData, isLoading, isError, error } = useJobProcessTree({
    cluster,
    jobId,
    client,
  })

  const nodeNames = useMemo(() => {
    if (!treeData?.nodes) return []
    return Object.keys(treeData.nodes)
  }, [treeData])

  const [selectedNode, setSelectedNode] = useState<string | null>(null)
  const activeNode = selectedNode ?? nodeNames[0] ?? null

  const activeNodeData = activeNode && treeData?.nodes ? treeData.nodes[activeNode] : null
  const metadata = activeNodeData?.metadata ?? null

  // Timeline state
  const sampleSteps = useMemo(() => {
    if (!showTimeline || !activeNodeData) return []
    return collectUniqueSampleTimes(activeNodeData)
  }, [showTimeline, activeNodeData])

  const [currentStepIndex, setCurrentStepIndex] = useState<number | null>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const playIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Initialize to last step when data loads or node changes
  useEffect(() => {
    if (sampleSteps.length > 0) {
      setCurrentStepIndex(sampleSteps.length - 1)
      setIsPlaying(false)
    }
  }, [sampleSteps])

  // Playback logic
  useEffect(() => {
    if (isPlaying && sampleSteps.length > 1) {
      playIntervalRef.current = setInterval(() => {
        setCurrentStepIndex((prev) => {
          const current = prev ?? 0
          if (current >= sampleSteps.length - 1) {
            setIsPlaying(false)
            return sampleSteps.length - 1
          }
          return current + 1
        })
      }, 150)
    }
    return () => {
      if (playIntervalRef.current) clearInterval(playIntervalRef.current)
    }
  }, [isPlaying, sampleSteps.length])

  const { flowNodes, flowEdges } = useMemo(() => {
    if (!activeNodeData) {
      return { flowNodes: [], flowEdges: [] }
    }

    let rawNodes, rawEdges
    const isLastStep = currentStepIndex !== null && currentStepIndex >= sampleSteps.length - 1
    if (showTimeline && sampleSteps.length > 0 && currentStepIndex !== null && !isLastStep) {
      const targetTime = sampleSteps[currentStepIndex]
      ;({ nodes: rawNodes, edges: rawEdges } = processTreeToFlowAtTime(activeNodeData, targetTime))
    } else {
      ;({ nodes: rawNodes, edges: rawEdges } = processTreeToFlow(activeNodeData))
    }

    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(rawNodes, rawEdges, direction)
    return { flowNodes: layoutedNodes, flowEdges: layoutedEdges }
  }, [activeNodeData, showTimeline, sampleSteps, currentStepIndex, direction])

  const [nodes, setNodes, onNodesChange] = useNodesState(flowNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(flowEdges)

  useEffect(() => {
    setNodes(flowNodes)
    setEdges(flowEdges)
    requestAnimationFrame(() => {
      fitView({ padding: 0.2 })
    })
  }, [flowNodes, flowEdges, setNodes, setEdges, fitView])

  const onLayout = useCallback((dir: string) => {
    setDirection(dir as 'TB' | 'LR')
  }, [])

  if (isLoading) {
    return (
      <Box w="100%" h="100%" display="flex" alignItems="center" justifyContent="center">
        <Spinner size="lg" />
      </Box>
    )
  }

  if (isError) {
    return (
      <Alert.Root status="error">
        <Alert.Indicator />
        <Alert.Description>
          Failed to load process tree: {error?.message || 'Unknown error'}
        </Alert.Description>
      </Alert.Root>
    )
  }

  if (!treeData || nodeNames.length === 0) {
    return (
      <Alert.Root status="info">
        <Alert.Indicator />
        <Alert.Description>No process tree data available for this job.</Alert.Description>
      </Alert.Root>
    )
  }

  const showSlider = showTimeline && sampleSteps.length > 1 && currentStepIndex !== null

  return (
    <VStack w="100%" h="100%" gap={0}>
      <Box w="100%" flex="1" minH={0}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          colorMode={colorMode}
          fitView
          nodesConnectable={false}
          nodesDraggable={true}
          elementsSelectable={true}
        >
          <Controls />
          <Panel position="top-right">
            <HStack gap={2}>
              {nodeNames.length > 1 && (
                <SegmentGroup.Root
                  size="sm"
                  value={activeNode ?? ''}
                  onValueChange={(e) => setSelectedNode(e.value)}
                >
                  <SegmentGroup.Indicator />
                  <SegmentGroup.Items
                    items={nodeNames.map((name) => ({ value: name, label: name }))}
                  />
                </SegmentGroup.Root>
              )}
              <SegmentGroup.Root defaultValue="TB" onValueChange={(e) => e.value && onLayout(e.value)}>
                <SegmentGroup.Indicator />
                <SegmentGroup.Items
                  items={[
                    { value: 'TB', label: <MdOutlineAlignVerticalTop /> },
                    { value: 'LR', label: <MdOutlineAlignHorizontalLeft /> },
                  ]}
                />
              </SegmentGroup.Root>
            </HStack>
          </Panel>
          <MiniMap />
          <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
        </ReactFlow>
      </Box>
      {showSlider && (
        <TimelineSlider
          steps={sampleSteps}
          currentStepIndex={currentStepIndex}
          onStepChange={setCurrentStepIndex}
          isPlaying={isPlaying}
          onPlayToggle={() => setIsPlaying((p) => !p)}
          processCount={flowNodes.length}
          totalProcesses={metadata?.total_processes ?? 0}
        />
      )}
    </VStack>
  )
})

ProcessTreeFlowInner.displayName = 'ProcessTreeFlowInner'

export const ProcessTreeView = memo(({ cluster, jobId, client, showTimeline }: ProcessTreeViewProps) => {
  return (
    <ReactFlowProvider>
      <ProcessTreeFlowInner cluster={cluster} jobId={jobId} client={client} showTimeline={showTimeline} />
    </ReactFlowProvider>
  )
})

ProcessTreeView.displayName = 'ProcessTreeView'

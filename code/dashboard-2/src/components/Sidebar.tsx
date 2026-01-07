import React from 'react'
import {
  createTreeCollection,
  TreeView,
  Box,
  Button,
  VStack,
} from '@chakra-ui/react'
import { NavLink, useLocation, useNavigate } from 'react-router'
import { LuChevronRight, LuPlus, LuGripVertical, LuSettings } from 'react-icons/lu'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

import { useCluster } from '../hooks/useCluster'
import { getClusterConfig, getClusterFullName, AVAILABLE_CLUSTERS } from '../config/clusters'

export const Sidebar = () => {
  const navigate = useNavigate()
  const { selectedClusters } = useCluster()
  
  // Check if all available clusters are already added
  const allClustersAdded = selectedClusters.length === AVAILABLE_CLUSTERS.length
  
  return (
    <VStack gap={0} h="100%" alignItems="stretch">
      <Box flex={1} overflowY="auto" overflowX="hidden" minH="0">
        <SideBarContent />
      </Box>
      <Box p={3} borderTopWidth="1px" flexShrink={0}>
        <Button
          size="sm"
          variant="outline"
          width="100%"
          onClick={() => navigate('/v2/select-cluster')}
          colorPalette="blue"
        >
          {allClustersAdded ? <LuSettings /> : <LuPlus />}
          {allClustersAdded ? 'Manage Clusters' : 'Add Cluster'}
        </Button>
      </Box>
    </VStack>
  )
}

// Sidebar items in tree collection format
interface TreeNode {
  id: string
  name: string
  path?: string
  matches?: string
  icon?: React.ComponentType
  children?: TreeNode[]
}

const transformSidebarItemsToTree = (selectedClusters: string[]) => {
  const nodes: TreeNode[] = []

  // Common sub-routes for cluster dashboards
  const CLUSTER_ROUTES = [
    { text: 'Overview', route: '/overview' },
    { text: 'Partitions', route: '/partitions' },
    { text: 'Nodes', route: '/nodes' },
    { text: 'Jobs', route: '/jobs' },
    { text: 'Queries', route: '/queries' },
    { text: 'Errors', route: '/errors' },
  ]

  // Build nodes for selected clusters only
  selectedClusters.forEach((clusterId) => {
    const config = getClusterConfig(clusterId)
    if (!config) return

    const clusterFullName = getClusterFullName(clusterId)
    const basePath = `/v2/${clusterFullName}`

    const node: TreeNode = {
      id: `cluster-${clusterId}`,
      name: config.name,
      path: `/dashboard/${clusterId}`,
      matches: clusterFullName,
      icon: config.icon,
      children: CLUSTER_ROUTES.map((route, subIndex) => ({
        id: `cluster-${clusterId}-sub-${subIndex}`,
        name: route.text,
        path: basePath + route.route,
        matches: clusterFullName + route.route,
      })),
    }

    nodes.push(node)
  })

  return createTreeCollection<TreeNode>({
    nodeToValue: (node) => node.id,
    nodeToString: (node) => node.name,
    rootNode: {
      id: 'ROOT',
      name: '',
      children: nodes,
    },
  })
}

const SideBarContent = () => {
  const location = useLocation()
  const {pathname} = location
  const { selectedClusters, reorderClusters } = useCluster()
  const [activeId, setActiveId] = React.useState<string | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string)
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event

    if (over && active.id !== over.id) {
      const oldIndex = selectedClusters.indexOf(active.id as string)
      const newIndex = selectedClusters.indexOf(over.id as string)

      const newOrder = [...selectedClusters]
      const [removed] = newOrder.splice(oldIndex, 1)
      newOrder.splice(newIndex, 0, removed)

      reorderClusters(newOrder)
    }

    setActiveId(null)
  }

  const collection = transformSidebarItemsToTree(selectedClusters)

  const getSelectedValue = () => {
    let selected: string[] = []
    let parentId: string | null = null
    let bestMatchLength = 0

    // Find all nodes and check for matches - prioritize longer matches
    const findSelected = (nodes: TreeNode[], parentNodeId?: string): void => {
      for (const node of nodes) {
        if (node.matches && pathname.includes(node.matches)) {
          // Select the node with the longest match for better specificity
          if (node.matches.length > bestMatchLength) {
            selected = [node.id]
            parentId = parentNodeId || null
            bestMatchLength = node.matches.length
          }
        }
        if (node.children) {
          findSelected(node.children, node.id)
        }
      }
    }

    // Get root node and search through children
    const rootNode = collection.rootNode as TreeNode
    if (rootNode.children) {
      findSelected(rootNode.children)
    }

    // If a child is selected, also include its parent
    if (parentId && selected.length > 0) {
      return [parentId, ...selected]
    }

    return selected
  }

  const activeCluster = activeId ? getClusterConfig(activeId) : null

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <SortableContext
        items={selectedClusters}
        strategy={verticalListSortingStrategy}
      >
        <TreeView.Root
          collection={collection}
          variant="subtle"
          size="md"
          colorPalette="blue"
          selectionMode="multiple"
          selectedValue={getSelectedValue()}
          defaultExpandedValue={collection.getBranchValues()}
          animateContent
        >
          <TreeView.Tree>
            <TreeView.Node
              indentGuide={<TreeView.BranchIndentGuide />}
              render={({ node, nodeState }) => {
                const treeNode = node as TreeNode
                const clusterId = treeNode.id.replace('cluster-', '')

                return nodeState.isBranch ? (
                  <SortableTreeBranch
                    clusterId={clusterId}
                    treeNode={treeNode}
                    nodeState={nodeState}
                  />
                ) : (
                  <TreeView.Item asChild>
                    <NavLink to={treeNode.path || '#'}>
                      <TreeView.ItemText>{treeNode.name}</TreeView.ItemText>
                    </NavLink>
                  </TreeView.Item>
                )
              }}
            />
          </TreeView.Tree>
        </TreeView.Root>
      </SortableContext>
      <DragOverlay>
        {activeId && activeCluster ? (
          <Box
            px={2}
            py={1}
            borderRadius="md"
            bg="blue.100"
            boxShadow="lg"
            display="flex"
            alignItems="center"
            fontWeight="semibold"
            width="200px"
          >
            <Box fontSize="md" mr={1}>
              <LuChevronRight />
            </Box>
            {activeCluster.icon && (
              <Box fontSize="lg" mr={2}>
                {React.createElement(activeCluster.icon)}
              </Box>
            )}
            <Box flex={1}>{activeCluster.name}</Box>
          </Box>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}

// Sortable wrapper for tree branch nodes
interface SortableTreeBranchProps {
  clusterId: string
  treeNode: TreeNode
  nodeState: {
    isBranch: boolean
    selected: boolean
    expanded: boolean
  }
}

const SortableTreeBranch: React.FC<SortableTreeBranchProps> = ({ 
  clusterId, 
  treeNode, 
  nodeState 
}) => {
  const [isHovered, setIsHovered] = React.useState(false)
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: clusterId })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    width: '100%',
  }

  return (
    <div 
      ref={setNodeRef} 
      style={style}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <TreeView.BranchControl
        fontWeight={nodeState.selected ? 'bold' : 'semibold'}
        display="flex"
        alignItems="center"
        position="relative"
        px={2}
        py={1}
        borderRadius="md"
        bg={!isDragging ? 'transparent' : 'transparent'}
        _hover={{ bg: 'gray.50' }}
        transition="all 0.2s"
        opacity={isDragging ? 0.5 : 1}
        width="100%"
      >
        <TreeView.BranchIndicator asChild>
          <Box fontSize="md" transition="transform 0.2s" mr={1}>
            <LuChevronRight />
          </Box>
        </TreeView.BranchIndicator>
        {treeNode.icon && (
          <Box fontSize="lg" mr={2}>
            {React.createElement(treeNode.icon)}
          </Box>
        )}
        <TreeView.BranchText flex={1}>{treeNode.name}</TreeView.BranchText>
        <Box
          {...attributes}
          {...listeners}
          cursor="grab"
          _active={{ cursor: 'grabbing' }}
          display="inline-flex"
          alignItems="center"
          
          p={1}
          borderRadius="sm"
          color={isHovered || isDragging ? 'blue.500' : 'transparent'}
          bg={isHovered && !isDragging ? 'blue.50' : 'transparent'}
          opacity={isHovered || isDragging ? 1 : 0}
          transform={isHovered || isDragging ? 'scale(1)' : 'scale(0.8)'}
          transition="all 0.15s ease-in-out"
          fontSize="lg"
          _hover={{ color: 'blue.600', bg: 'blue.100' }}
        >
          <LuGripVertical />
        </Box>
      </TreeView.BranchControl>
    </div>
  )
}

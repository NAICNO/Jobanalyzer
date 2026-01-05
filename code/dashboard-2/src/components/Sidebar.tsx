import React from 'react'
import {
  createTreeCollection,
  TreeView,
} from '@chakra-ui/react'
import { NavLink, useLocation } from 'react-router'
import { LuChevronRight } from 'react-icons/lu'

import { SIDEBAR_ITEMS } from '../Constants.ts'

export const Sidebar = () => {
  return <SideBarContent />
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

const transformSidebarItemsToTree = () => {
  const nodes: TreeNode[] = []

  SIDEBAR_ITEMS.forEach((item, index) => {
    if (item.type === 'separator') {
      return
    }

    const node: TreeNode = {
      id: `item-${index}`,
      name: item.text || '',
      path: item.path,
      matches: item.matches,
      icon: item.icon,
    }

    if (item.subItems && item.subItems.length > 0) {
      node.children = item.subItems.map((subItem, subIndex) => ({
        id: `item-${index}-sub-${subIndex}`,
        name: subItem.text,
        path: subItem.path,
        matches: subItem.matches,
      }))
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

  const collection = transformSidebarItemsToTree()

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

  return (
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

            return nodeState.isBranch ? (
              <TreeView.BranchControl
                fontWeight={nodeState.selected ? 'bold' : 'semibold'}
              >
                <TreeView.BranchIndicator asChild>
                  <LuChevronRight />
                </TreeView.BranchIndicator>
                {treeNode.icon && React.createElement(treeNode.icon)}
                <TreeView.BranchText>{treeNode.name}</TreeView.BranchText>
              </TreeView.BranchControl>
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
  )
}

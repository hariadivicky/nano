package nano

import "strings"

// node defines tree node.
type node struct {
	urlPattern string
	urlPart    string
	childrens  []*node
	isWildcard bool
}

// insertChildren functions to insert node as children.
// this function will call recursively as length of urlParts and cursor position (level)
func (n *node) insertChildren(urlPattern string, urlParts []string, level int) {

	// last inserted node cause cursor (level) has reached maximum value.
	// will stop recursive calls.
	if len(urlParts) == level {
		// fill url pattern to marks current node as complete url pattern.
		n.urlPattern = urlPattern

		return
	}

	urlPart := urlParts[level]

	// scan existance of current url part in children list.
	child := n.findChildren(urlPart)
	if child == nil {
		// current url part is not already registered as children node.
		// register children now.
		isWildcard := urlPart[0] == ':' || urlPart[0] == '*'
		child = &node{urlPart: urlPart, isWildcard: isWildcard}
		n.childrens = append(n.childrens, child)
	}

	// insert next urlParts as next level children.
	// moving cursor to next urlParts.
	child.insertChildren(urlPattern, urlParts, level+1)
}

// findChildren functions to find children by url part value.
// this function may return nil value.
func (n *node) findChildren(urlPart string) *node {

	// scanning for children
	for _, child := range n.childrens {
		// if current child url part is match or contain wildcard, so it's found.
		if child.urlPart == urlPart || child.isWildcard {
			return child
		}
	}

	// there is no children with current urlPart
	return nil
}

// findNode functions to find node.
// first (n *node) may be node that located at router.nodes[requestMethod].
func (n *node) findNode(searchParts []string, level int) *node {
	// cursor (level) reached maximum position.
	// or current url part has * wildcard
	if len(searchParts) == level || strings.HasPrefix(n.urlPart, "*") {
		// if current pattern has no url pattern, this mean current node doesn't complete.
		// not found.
		if n.urlPattern == "" {
			return nil
		}

		return n
	}

	// get current search part by cursor (level).
	urlPart := searchParts[level]

	// scan for nested childrens*.
	// *please read about getChildren.
	for _, child := range n.getChildren(urlPart) {
		// move cursor, scan recursively.
		result := child.findNode(searchParts, level+1)
		// found!
		if result != nil {
			return result
		}

		return nil
	}

	return nil
}

// getChildren functions to find children that has certain part
// or it's a wildcard
func (n *node) getChildren(urlPart string) []*node {
	nodes := make([]*node, 0)

	for _, node := range n.childrens {
		if node.urlPart == urlPart || node.isWildcard {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

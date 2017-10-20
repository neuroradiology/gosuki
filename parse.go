package main

import (
	"fmt"
	"encoding/json"
	"bytes"
)

type Node struct {
	Type string
	Children []interface{}
	Url string `json:",omitempty"`
	Name string
}


type RootData struct {
	Name string
	Roots map[string]Node
	Version float64
}



func mapToNode(childNode interface{}) (*Node, error) {
	if childNode == nil {
		return new(Node), nil
	}

	buf := new(bytes.Buffer)

	// Convert interface{} to json
	err := json.NewEncoder(buf).Encode(childNode)
	if err != nil {
		return nil, err
	}

	//fmt.Println(buf)

	out := new(Node)

	// Convert json to Node struct
	err = json.NewDecoder(buf).Decode(out)
	if err != nil {
		return nil, err
	}

	return out, nil
}



func parseJsonNodes(node *Node) {

	//fmt.Println("parsing node ", node.Name)

	if (node.Type == "url") {
		fmt.Println(node.Url)
	} else if (len(node.Children) != 0) { // If node is Folder
		for _, _childNode := range node.Children {
			// Type of childNode is interface{}
			//childNode := Node{}
			childNode, err := mapToNode(_childNode)
			if err != nil {
				panic(err)
			}
			parseJsonNodes(childNode)
		}
	}

	return
}

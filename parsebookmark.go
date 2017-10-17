package main

import "fmt"
import "io/ioutil"
import "encoding/json"
import "bytes"


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



func parse(node *Node) {

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
			parse(childNode)
		}
	}

	return
}

func main() {


	f, _ := ioutil.ReadFile("test.json")

	//var data Data // does not allocate memory (must use make)

	rootData := RootData{}
	_ = json.Unmarshal(f, &rootData)


	//fmt.Println(rootData)

	//fmt.Printf("Value: %v\nType: %T\n", rootData.Roots, rootData.Roots)

	for _, root := range rootData.Roots {
		parse(&root)
	}

}


/// Exercise

// - 1 Load people from json , load embedded objects
// - 2 convert date_added in bookmarks to same as mozilla bookmarks (create a go function) // MSDN FILETIME

//https://stackoverflow.com/questions/19074423/how-to-parse-the-date-added-field-in-chrome-bookmarks-file
//https://cs.chromium.org/chromium/src/base/time/time_win.cc?sq=package:chromium&type=cs
//https://stackoverflow.com/questions/6161776/convert-windows-filetime-to-second-in-unix-linux




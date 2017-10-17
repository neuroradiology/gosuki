package main

import "fmt"
import "encoding/json"

func main() {

	byt := []byte(`{"num":6,"strs":["à","b"]}`)

	var data map[string]interface{}

	if err := json.Unmarshal(byt, &data); err != nil {
		panic(err)
	}

	//fmt.Println(data)

	num := data["num"]
	fmt.Printf("Value: %v\nType: %T\n", num, num)
	strs := data["strs"].([]interface{})
	fmt.Printf("Value: %v\nType: %T\n", strs[0], strs[0])



}

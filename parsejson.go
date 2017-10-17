package main

import "fmt"
import "io/ioutil"
import "encoding/json"

type Person struct {
	Name string
	Age int
}

type Data struct {
	Name string
	Url string
	Tags []string `json:"categories"`
	//People []Person
}

func main() {


	f, _ := ioutil.ReadFile("test2.json")

	//var data Data // does not allocate memory (must use make)
	data := Data{}


	err := json.Unmarshal(f, &data)

	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println(data["people"])
	fmt.Println(data)

}


/// Exercise

// - 1 Load people from json , load embedded objects
// - 2 convert date_added in bookmarks to same as mozilla bookmarks (create a go function) // MSDN FILETIME

//https://stackoverflow.com/questions/19074423/how-to-parse-the-date-added-field-in-chrome-bookmarks-file
//https://cs.chromium.org/chromium/src/base/time/time_win.cc?sq=package:chromium&type=cs
//https://stackoverflow.com/questions/6161776/convert-windows-filetime-to-second-in-unix-linux




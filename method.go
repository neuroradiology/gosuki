package main

import "fmt"
import "strings"
import "math"


type MyName string

func printName(name MyName) {
	fmt.Println(name)
}

func camelCase(name MyName) {
	result := strings.Split(string(name), " ")

	//fmt.Println(result[0][0])

	firstName := strings.ToUpper(string(result[0][0])) + result[0][1:]
	lastName  := strings.ToUpper(string(result[1][0])) + result[1][1:]

	fmt.Printf("First Name: %s\nLast Name: %s", firstName, lastName)
}




type Rectangle struct {
	width float64
	height float64
	name string
}

type Triangle struct {
	height float64
	base float64
	name string
}

type Circle struct {
	radius float64
	circleName string
}


func (c *Circle) Area() float64 {
	return math.Pow(c.radius, 2) * math.Pi
}

func (c *Circle) Name() string {
	return c.circleName
}


func (t *Triangle) Area() float64 {
	return ( t.height  * t.base ) / 2
}

func (t *Triangle) Name() string {
	return t.name
}

func (r *Rectangle) Name() string {
    return r.name
}

func (r *Rectangle) Area() float64 {
	return r.width * r.height }

func (r *Rectangle) Scale(amount float64) {
	r.width = r.width * amount
	r.height = r.height * amount
}

type Geometric interface {
	Area() float64
	Name() string
}

type myInt int

func (i *myInt) Area() float64 {
	return 0.5
}

func (i *myInt) Name() string {
	return "I am an int"
}


func printArea(g []Geometric) {
	for _, obj := range(g)  {
		fmt.Printf("%v ---> Area: %v\n", obj.Name(), obj.Area())
	}
}



func main() {


	r := Rectangle{2,5, "rectangle"}
	t := Triangle{3, 5, "triangle"}
	c := Circle{2, "circle"}
	var i myInt = 4

	g := []Geometric{&r, &t, &c, &i}


	printArea(g)



	//fmt.Printf("Rectangle: %v ---> Area: %v\n", r, r.Area())
	//fmt.Printf("Triangle: %v ---> Area: %v\n", t, t.Area())
	//fmt.Printf("Circle: %v ---> Area: %v\n", c, c.Area())

	//g = &r
	//printArea(g)

	//g = &t
	//printArea(g)

	//g = &c
	//printArea(g)

}

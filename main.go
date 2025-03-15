package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

// group A types | non-Pointers -> strings, ints, bools, floats, array, structs
// group B types | Pointers     -> slices, maps, functions

func circleArea(r float64) float64 {
	return math.Pi * r * r
}

func call(n string) {
	fmt.Printf("Called func call with param %v \n", n)
}

func cycleFunc(n []string, f func(string)) {
	for _, v := range n {
		f(v)
	}
}

func getInitials(n string) (string, string) {
	s := strings.ToUpper(n)
	names := strings.Split(s, " ")

	var initials []string
	for _, v := range names {
		initials = append(initials, v[:1])
	}

	if len(initials) > 1 {
		return initials[0], initials[1]
	}

	return initials[0], "_"
}

func pointerDecode(x *string) {
	*x = "cloud"
	fmt.Println(*x)
}

func main() {
	var greeting string = "Hello World!"
	greeting2 := "Hello World2!" // cannot reside outside func
	fmt.Println(greeting, greeting2)

	var number1 int = 1
	var number1Bit int16 = 225
	var numnerUnsigned uint16 = 10 // positive only
	fmt.Println(number1, number1Bit, numnerUnsigned)

	var decimal1 float32 = 0.1
	fmt.Println(decimal1)

	fmt.Println("Greeting is", greeting, "age is", number1Bit)
	fmt.Printf("Greeting is %v age is %v\n", greeting, number1Bit)

	var stringConcat string = fmt.Sprintf("Greeting is %v age is %v\n", greeting, number1Bit)
	fmt.Println(stringConcat)

	var ages = [3]int{1, 2, 3}
	fmt.Println(ages)

	var nums = []int{1, 2, 3}
	nums[1] = 2
	nums = append(nums, 4)

	var strings = []string{"1txt", "2txt", "3txt"}

	var rangenum = ages[1:2]

	fmt.Println(nums, rangenum)

	menu := map[string]float64{
		"soup":           4.99,
		"pie":            7.99,
		"salad":          6.99,
		"toffee pudding": 3.55,
	}

	fmt.Println(menu["soup"])

	for k, v := range menu {
		fmt.Printf("The price of %v is %v\n", k, v)
	}

	//

	for i := 0; i < len(strings); i++ {
		fmt.Println(strings[i])
	}

	for index, value := range strings {
		if index == 0 {
			fmt.Printf("Index is in %v", index)
		}
		fmt.Println(index, value)
	}

	if len(nums) > 0 {
		fmt.Printf("nums length is %v", len(nums))
	}

	cycleFunc([]string{"one", "two", "three"}, call)
	circleArea(10.1)
	fn1, sn1 := getInitials("Edcel Vista")
	fmt.Printf("%v %v \n", fn1, sn1)

	r := sayHello("Edcel")
	fmt.Printf("%v Edcel!\n", r)

	name := "tifa"
	fmt.Println("Memory Address of name is:", &name)

	m := &name
	fmt.Println("Memory Address of name is:", m)
	fmt.Println("The value at Memory Address of name is:", *m)

	pointerDecode(m)

	fmt.Println(name)

	//

	myBill := newBill("Edcel's Bill")
	myBill.updateTip(10)
	myBill.addItem("bread", 1.99)

	fmt.Println(myBill.format())

	myNewBill := createBill()
	promtOptions(myNewBill)

	x := map[string]string{
		"foo": "bar",
	}

	data, _ := json.Marshal(x)
	fmt.Println(string(data))

	data2 := []byte(`{"foo":"bar"}`)
	var x2 person
	_ = json.Unmarshal(data2, &x2)
	fmt.Println(x)

}

type person struct {
	Foo string `json: foo`
}

func getInput(prompt string, r *bufio.Reader) (string, error) {
	fmt.Print(prompt)
	input, err := r.ReadString('\n')

	return strings.TrimSpace(input), err
}

func promtOptions(b bill) {
	reader := bufio.NewReader(os.Stdin)
	opt, _ := getInput("Choose Option (a - add item, s - save the bill, t - add a tip): ", reader)
	fmt.Println(opt)
}

func createBill() bill {
	reader := bufio.NewReader(os.Stdin)
	name, _ := getInput("Create a new bill name: ", reader)

	b := newBill(name)
	fmt.Println("Created the bill - ", b.name)

	return b
}

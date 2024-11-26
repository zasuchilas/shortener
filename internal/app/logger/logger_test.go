package logger

import "fmt"

func ExampleInitialize() {
	err := Initialize("debug")
	fmt.Println(err)

	err2 := Initialize("wrong")
	fmt.Println(err2.Error())

	// Output:
	// <nil>
	// unrecognized level: "wrong"
}

func ExampleServiceInfo() {
	ServiceInfo("1.2.3")

	// Output:
	//
}

func ExampleConfigInfo() {
	ConfigInfo()

	// Output:
	//
}

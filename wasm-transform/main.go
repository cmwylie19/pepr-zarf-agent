// package main

// import (
// 	"syscall/js"
// )

// func add(this js.Value, args []js.Value) interface{} {
// 	if len(args) != 2 {
// 		return nil
// 	}
// 	return args[0].Int() + args[1].Int()
// }

// func main() {
// 	js.Global().Set("add", js.FuncOf(add))
// 	select {}
// }

package main

import (
	"syscall/js"

	"github.com/defenseunicorns/zarf/src/pkg/transform"
)

func WasmImageTransformHost(this js.Value, args []js.Value) (string, error) {
	targetHost := args[0].string()
	srcReference := args[1].string()
	transformedImage, _ := transform.ImageTransformHost(targetHost, srcReference)
	return transformedImage, nil
}

func main() {
	c := make(chan bool)
	js.Global().Set("WasmImageTransformHost", js.FuncOf(WasmImageTransformHost))
	<-c
}

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

//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/defenseunicorns/zarf/src/pkg/transform"
)

func ImageTransformHost(this js.Value, args []js.Value) interface{} {
	targetHost := args[0].String()
	srcReference := args[1].String()
	transformedImage, _ := transform.ImageTransformHost(targetHost, srcReference)
	return transformedImage
}

func main() {
	c := make(chan bool)
	js.Global().Set("WasmImageTransformHost", js.FuncOf(ImageTransformHost))
	<-c
}

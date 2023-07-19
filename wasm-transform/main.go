package main

import (

	// "syscall/js"

	// "fmt"

	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/defenseunicorns/zarf/src/pkg/transform"
	admission "k8s.io/api/admission/v1"
)

// WASMRequest will be create in Pepr Admission Controller
//
//	type WASMRequest[T any] struct {
//		Request *admission.AdmissionReview `json:"request,omitempty" protobuf:"bytes,1,opt,name=request"`
//		// Kubernetes Resource
//		Resource T
//		// Function Arguments
//		Args []interface{}
//	}

// Convert js.Value to WASM Request
type WASMRequest struct {
	Request admission.AdmissionReview `json:"request,omitempty" protobuf:"bytes,1,opt,name=request"`
	// Kubernetes Resource
	// Resource T
	// Function Arguments
	Args []interface{} `json:"args,omitempty" protobuf:"bytes,1,opt,name=args"`
}

//	type WASMRequest struct {
//		Egress map[string]interface{}
//	}
type WASMRequestIngress struct {
}
type WASMRequestEgress struct {
	Request  []byte        // Represents admission.AdmissionReview
	Resource []byte        // Represents a Kubernetes Resources
	Args     []interface{} // Represents arguments
}

var wasmRequest WASMRequest
var wasmRequestEgress WASMRequestEgress

// argument is type of resource
// func ConvertResource(this js.Value, args []js.Value) interface{} {

// }
func GetResource(this js.Value, args []js.Value) interface{} {
	return fmt.Sprintf("%s", args[0])
}
func TransformImage(this js.Value, args []js.Value) interface{} {
	targetHost := fmt.Sprintf("%s", args[0])
	srcReference := fmt.Sprintf("%s", args[1])
	transformed, _ := transform.ImageTransformHost(targetHost, srcReference)
	wasmRequest.Args = []interface{}{targetHost, srcReference}
	return transformed
}
func ConvertAdmissionRequest(this js.Value, args []js.Value) interface{} {
	var admission admission.AdmissionReview

	// Convert AdmissionReview js.Value to bytes
	admissionString := []byte(fmt.Sprintf("%s", args[0].String()))

	// Unmarshall into AdmissionReview object
	if err := json.Unmarshal([]byte(admissionString), &admission); err != nil {
		fmt.Println("could not unmarshall argument to WASMRequest - RawReview ", err)
		return nil
	}

	// Build WASM Request
	wasmRequest.Request = admission

	// Marshall wasmRequest into a string - to send back
	marshalledString, err := json.Marshal(wasmRequest.Request)
	if err != nil {
		fmt.Println("Couldnt marshal the string")
	}

	// Create an empty interface to hold the parsed JSON data
	var data interface{}

	// Unmarshal the JSON string into the interface
	err = json.Unmarshal([]byte(marshalledString), &data)
	if err != nil {
		fmt.Println("Error:", err)

	}

	// Marshal the interface into a pretty-printed JSON string
	wasmRequestEgress.Request, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)

	}

	//return string(wasmRequestEgress.Request)
	foods := map[string]interface{}{
		"bacon": "delicious",
		"eggs": map[string]interface{}{
			"chicken": 1.75,
		},
		"steak": true,
	}
	return foods
	//return string(wasmRequestEgress.Request)
	//return
	//return wasmRequest

	// QuickType used to generate Go structs to TS Interfaces
	// Play with a map of interfaces
}
func add(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return nil
	}
	result := args[0].Int() + args[1].Int()
	fmt.Println(result)
	return result
}

func main() {

	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("ConvertAdmissionRequest", js.FuncOf(ConvertAdmissionRequest))
	js.Global().Set("TransformImage", js.FuncOf(TransformImage))
	js.Global().Set("GetResource", js.FuncOf(GetResource))
	// research how to scope functions to a module
	// - would  all wasm functions from an AdmissionPatch
	// -- pod.Merge and return a deepPartial from WASM
	// use reflect to map to string interface
	// we could do all annotations

	select {}
}

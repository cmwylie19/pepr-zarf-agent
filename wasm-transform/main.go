package main

import (

	// "syscall/js"

	// "fmt"

	"fmt"
	"syscall/js"
	"unicode/utf8"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v2"
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
const RawReview = `
apiVersion: admission.k8s.io/v1
kind: AdmissionReview
request:
  uid: 705ab4f5-6393-11e8-b7cc-42010a800002
  kind:
    group: autoscaling
    version: v1
    kind: Scale
  resource:
    group: apps
    version: v1
    resource: deployments
  subResource: scale

  requestKind:
    group: autoscaling
    version: v1
    kind: Scale
  requestResource:
    group: apps
    version: v1
    resource: deployments
  requestSubResource: scale
  name: my-deployment
  namespace: my-namespace
  operation: UPDATE

  userInfo:
    username: admin

    uid: 014fbff9a07c
    groups:
      - system:authenticated
      - my-admin-group
    extra:
      some-key:
        - some-value1
        - some-value2
  object:
    apiVersion: autoscaling/v1
    kind: Scale
  oldObject:
    apiVersion: autoscaling/v1
    kind: Scale
  options:
    apiVersion: meta.k8s.io/v1
    kind: UpdateOptions
  dryRun: False
`

type WASMRequest struct {
	Request admission.AdmissionReview `json:"request,omitempty" protobuf:"bytes,1,opt,name=request"`
	// Kubernetes Resource
	// Resource T
	// Function Arguments
	Args []interface{} `json:"args,omitempty" protobuf:"bytes,1,opt,name=args"`
}

func MarshalReview() admission.AdmissionReview {
	var admission admission.AdmissionReview

	if err := yaml.Unmarshal([]byte(RawReview), &admission); err != nil {
		fmt.Errorf("could not unmarshall argument to WASMRequest ", err)
	}
	return admission
}

// PrintDiff prints the differences between a and b with a as original and b as new
func PrintDiff(textA, textB string) {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(textA, textB, true)

	diffs = dmp.DiffCleanupSemantic(diffs)

	fmt.Println(dmp.DiffPrettyHtml(diffs))
}
func ConvertToWASMRequest(this js.Value, args []js.Value) interface{} {
	wasmRequest := WASMRequest{}
	var admission admission.AdmissionReview
	fmt.Printf("%s", args[0].String())

	if utf8.Valid([]byte(args[0].String())) {
		fmt.Println("Valid")
	} else {
		fmt.Println("Not valid")
	}
	fmt.Printf("Type %T", args[0].String())
	admissionString := []byte(fmt.Sprintf("%s", args[0].String()))

	// compare strings
	if fmt.Sprintf("%s", admissionString) == RawReview {
		fmt.Println("They are equal")
	} else {
		fmt.Println("They are not equal")
		// PrintDiff(admissionString, RawReview)
		// fmt.Println(admissionString + "\n----\n" + RawReview)

	}

	for i, val := range []byte(RawReview) {
		if admissionString[i] != val {
			fmt.Printf("False! Bytes different %d %s != %s", i, string(admissionString[i]), string(val))
			break
		}
	}

	if err := yaml.Unmarshal([]byte(RawReview), &admission); err != nil {
		fmt.Println("could not unmarshall argument to WASMRequest - RawReview ", err)
		return nil
	}

	// bytes := []byte(admissionString)
	// fmt.Printf("Value of bytes is %v", reflect.bytes)
	if err := yaml.Unmarshal(admissionString, &admission); err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("Got to the end")
	wasmRequest.Request = admission

	marshalledString, err := yaml.Marshal(wasmRequest.Request)
	if err != nil {
		fmt.Println("Couldnt marshal the string")
	}

	// jsonData, err := y.YAMLToJSON([]byte(marshalledString))
	// if err != nil {
	// 	fmt.Println("Error marshaling JSON:", err)
	// 	return nil
	// }

	return string(marshalledString)
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
	// wasmReq := SendWASMRequest()
	// fmt.Printf("%s", fmt.Sprintf("%s", wasmReq))
	fmt.Printf("%+v", MarshalReview())
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("ConvertToWASMRequest", js.FuncOf(ConvertToWASMRequest))

	select {}
}

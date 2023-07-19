package main

import (

	// "syscall/js"

	// "fmt"
	"encoding/json"
	"fmt"
	"log"
	"syscall/js"

	corev1 "k8s.io/api/core/v1"

	"github.com/defenseunicorns/zarf/src/pkg/transform"
	admission "k8s.io/api/admission/v1"
)

const AgentErrImageSwap = "Unable to swap the host for (%s)"
const replaceOperation = "replace"

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	From  string      `json:"from,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// ReplacePatchOperation returns a replace JSON patch operation.
func ReplacePatchOperation(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Op:    replaceOperation,
		Path:  path,
		Value: value,
	}
}

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

func podTransform(this js.Value, args []js.Value) interface{} {
	peprRequest := args[0].String()
	imagePullSecretName := args[1].String()
	targetHost := args[2].String()
	// srcReference := args[3].String()
	pod := &corev1.Pod{}
	var patchOperations []PatchOperation

	// Define a variable to hold the parsed JSON data
	var data map[string]interface{}

	// Unmarshal the JSON string into the data variable
	err := json.Unmarshal([]byte(peprRequest), &data)
	if err != nil {
		log.Fatal(err)
	}
	// Convert the interface to a JSON byte array
	podBytes, err := json.Marshal(data["Raw"])
	if err != nil {
		log.Fatal(err)
	}

	// unmarshal podBytes into pod
	err = json.Unmarshal(podBytes, pod)
	if err != nil {
		fmt.Println("Error unmarshalling pod", err)
	}

	// don't do anything if pod has ignore labels
	if checkIgnoreLabels(pod) {
		fmt.Println("Pod has ignore labels, ignoring")
		return nil
	}

	patchOperations = addImagePullSecret(pod, imagePullSecretName, patchOperations)
	patchOperations = transformContainerImages(pod, targetHost, patchOperations)
	patchOperations = addPatchedLabel(pod, patchOperations)

	// fmt.Printf("%s\n", pod.Name)
	fmt.Printf("POD\n%+v", pod)
	fmt.Println("Raw Object:\n", fmt.Sprintf("%s", data["Raw"]))
	// PrettyPrint
	// Create an empty interface to unmarshal the JSON string
	// Marshal the Pod object into a pretty printed JSON string
	podBytes, err = json.MarshalIndent(pod, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	// Convert the JSON bytes to a string
	podString := string(podBytes)
	return string(podString)
}
func addPatchedLabel(pod *corev1.Pod, patchOperations []PatchOperation) []PatchOperation {
	pod.Labels["zarf-agent"] = "patched"
	return append(patchOperations, ReplacePatchOperation("/metadata/labels/zarf-agent", "patched"))
}
func checkIgnoreLabels(pod *corev1.Pod) bool {
	// check if pod has ignoreLables
	if pod.Labels["zarf-agent"] == "patched" || pod.Labels["zarf.dev/agent"] == "ignore" {
		// We've already played with this pod, just keep swimming 🐟
		return true
	}
	return false
}

func addImagePullSecret(pod *corev1.Pod, imagePullSecretName string, patchOperations []PatchOperation) []PatchOperation {
	pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, corev1.LocalObjectReference{
		Name: imagePullSecretName,
	})
	return append(patchOperations, ReplacePatchOperation("/spec/imagePullSecrets", imagePullSecretName))
}

func transformContainerImages(pod *corev1.Pod, targetHost string, patchOperations []PatchOperation) []PatchOperation {
	// update the image host for each init container
	for idx, container := range pod.Spec.InitContainers {
		path := fmt.Sprintf("/spec/initContainers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(targetHost, container.Image)

		if err != nil {
			fmt.Printf(AgentErrImageSwap, err)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		pod.Spec.InitContainers[idx].Image = replacement
		patchOperations = append(patchOperations, ReplacePatchOperation(path, replacement))
	}

	// update the image host for each ephemeral container
	for idx, container := range pod.Spec.EphemeralContainers {
		path := fmt.Sprintf("/spec/ephemeralContainers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(targetHost, container.Image)
		if err != nil {
			fmt.Printf(AgentErrImageSwap, err)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		pod.Spec.EphemeralContainers[idx].Image = replacement
		patchOperations = append(patchOperations, ReplacePatchOperation(path, replacement))
	}

	// update the image host for each normal container
	for idx, container := range pod.Spec.Containers {
		path := fmt.Sprintf("/spec/containers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(targetHost, container.Image)
		if err != nil {
			fmt.Printf(AgentErrImageSwap, err)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		fmt.Println("replacement", replacement)
		pod.Spec.Containers[idx].Image = replacement
		fmt.Println(pod.Spec.Containers[idx].Image)
		patchOperations = append(patchOperations, ReplacePatchOperation(path, replacement))
	}
	return patchOperations
}
func main() {

	js.Global().Set("podTransform", js.FuncOf(podTransform))
	js.Global().Set("TransformImage", js.FuncOf(TransformImage))
	// research how to scope functions to a module
	// - would  all wasm functions from an AdmissionPatch
	// -- pod.Merge and return a deepPartial from WASM
	// use reflect to map to string interface
	// we could do all annotations

	select {}
}

// https://go.dev/play/p/AIzdixvCzNW

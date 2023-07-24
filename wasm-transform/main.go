package main

import (
	"encoding/json"
	"fmt"
	"log"
	"syscall/js"

	corev1 "k8s.io/api/core/v1"

	"github.com/defenseunicorns/zarf/src/pkg/transform"
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

// func ConrtAdmissionRequest(this js.Value, args []js.Value) interface{} {

// 	foods := map[string]interface{}{
// 		"bacon": "delicious",
// 		"eggs": map[string]interface{}{
// 			"chicken": 1.75,
// 		},
// 		"steak": true,
// 	}
// 	return foods
// 	//return string(wasmRequestEgress.Request)
// 	//return
// 	//return wasmRequest

// 	// QuickType used to generate Go structs to TS Interfaces
// 	// Play with a map of interfaces
// }

func podTransform(this js.Value, args []js.Value) interface{} {

	// Get arguments from Pepr
	rawRequest := args[0].String()
	// admissionRequest := args[1].String()
	imagePullSecretName := args[2].String()
	targetHost := args[3].String()

	pod := &corev1.Pod{}

	// Define a variable to hold the parsed JSON data
	var data map[string]interface{}

	// Unmarshal the JSON string into the data variable
	err := json.Unmarshal([]byte(rawRequest), &data)
	if err != nil {
		log.Fatal(err)
	}
	// Convert the interface to a JSON byte array
	podBytes, err := json.Marshal(data)
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

	addImagePullSecret(pod, imagePullSecretName)
	transformContainerImages(pod, targetHost)
	addPatchedLabel(pod)

	// fmt.Printf("%s\n", pod.Name)
	fmt.Printf("POD\n%+v", pod)

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
func addPatchedLabel(pod *corev1.Pod) {
	pod.Labels["zarf-agent"] = "patched"
}
func checkIgnoreLabels(pod *corev1.Pod) bool {
	// check if pod has ignoreLables
	if pod.Labels["zarf-agent"] == "patched" || pod.Labels["zarf.dev/agent"] == "ignore" {
		// We've already played with this pod, just keep swimming 🐟
		return true
	}
	return false
}

func addImagePullSecret(pod *corev1.Pod, imagePullSecretName string) {
	pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, corev1.LocalObjectReference{
		Name: imagePullSecretName,
	})
}

func transformContainerImages(pod *corev1.Pod, targetHost string) {
	// update the image host for each init container
	for idx, container := range pod.Spec.InitContainers {
		replacement, err := transform.ImageTransformHost(targetHost, container.Image)

		if err != nil {
			fmt.Printf(AgentErrImageSwap, err)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		pod.Spec.InitContainers[idx].Image = replacement
	}

	// update the image host for each ephemeral container
	for idx, container := range pod.Spec.EphemeralContainers {
		replacement, err := transform.ImageTransformHost(targetHost, container.Image)
		if err != nil {
			fmt.Printf(AgentErrImageSwap, err)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		pod.Spec.EphemeralContainers[idx].Image = replacement
	}

	// update the image host for each normal container
	for idx, container := range pod.Spec.Containers {
		replacement, err := transform.ImageTransformHost(targetHost, container.Image)
		if err != nil {
			fmt.Printf(AgentErrImageSwap, err)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		fmt.Println("replacement", replacement)
		pod.Spec.Containers[idx].Image = replacement
		fmt.Println(pod.Spec.Containers[idx].Image)
	}
}
func main() {
	done := make(chan struct{})
	//js.Global().Set("podTransform", js.FuncOf(podTransform))
	// js.Function

	js.Global().Set("zarfTransform", make(map[string]interface{}))
	module := js.Global().Get("zarfTransform")
	module.Set("podTransform", js.FuncOf(podTransform))
	<-done
	//select {}
}

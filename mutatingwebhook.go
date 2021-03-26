package main

import (
	"context"
	"encoding/json"
	"fmt"
	//"log"
	"net/http"

	srev1 "autopatch/api/v1"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	libjq "github.com/snowcrystall/jq-go"
)

// +kubebuilder:webhook:path=/wangjl-autopatch-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=auto-patch

// podAnnotator annotates Pods
type podAnnotator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func patchPod(patch string, pod *corev1.Pod) error {
	marshaledPod, err := libjq.Apply(patch, pod)
	fmt.Printf("modifiedby new pod json: %s\n", marshaledPod)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(marshaledPod[0], &pod); err != nil {
		return err
	}
	return nil
}

func (a *podAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	if pod.Labels["run"] == "" {
		return admission.Allowed("")
	}

	patchList := srev1.PatchList{}

	//newPod := corev1.Pod{}
	for k, v := range pod.Labels {
		if err := a.Client.List(ctx, &patchList, client.MatchingLabels{k: v}); err == nil {
			fmt.Printf("found patch with label: %s=%s\n", k, v)
			for _, patch := range patchList.Items {
				pod.Annotations["patched-by-"+patch.Name] = "true"
				for _, patch_step := range patch.Steps {
					if err := patchPod(patch_step.Patch, pod); err != nil {
						return admission.Errored(http.StatusInternalServerError, err)
					}

				}
			}
		} else {
			fmt.Printf("no found patch with label: %s=%s\n", k, v)
		}
	}

	//pod.Annotations["autopatch-webhook"] = "wangjl"
	//i, j := 0, 0
	//set_request_mem := true
	//for i = 0; i < len(pod.Spec.Containers); i++ {
	//	for j = 0; j < len(pod.Spec.Containers[i].Env); j++ {
	//		if pod.Spec.Containers[i].Env[j].Name == "REQUEST_MEM" {
	//			set_request_mem = false
	//			break
	//		}
	//	}
	//	if set_request_mem == true {
	//		fmt.Printf("add set resource env\n")
	//		request_mem_from := corev1.EnvVarSource{ResourceFieldRef: &corev1.ResourceFieldSelector{ContainerName: pod.Spec.Containers[i].Name, Resource: "requests.memory"}}
	//		request_mem_env := corev1.EnvVar{Name: "REQUEST_MEM", ValueFrom: &request_mem_from}
	//		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, request_mem_env)
	//		fmt.Printf("container env: %s\n", pod.Spec.Containers[i].Env)
	//	}
	//}

	//marshaledPod, err := json.Marshal(pod)
	//fmt.Printf("modifiedby new pod: %s\n", marshaledPod)
	//if err != nil {
	//	return admission.Errored(http.StatusInternalServerError, err)
	//}

	//marshaledPod2, err := libjq.Apply(".annotations[\"patch-engine\"]=\"jq\"", marshaledPod)
	//marshaledPod2, err := libjq.Apply(".metadata.annotations[\"patch-engine\"]=\"jq\"", pod)
	//fmt.Printf("modifiedby new pod2: %s\n", marshaledPod2)
	//if err != nil {
	//	return admission.Errored(http.StatusInternalServerError, err)
	//}

	//return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (a *podAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

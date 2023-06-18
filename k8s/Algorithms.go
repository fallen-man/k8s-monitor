package k8s

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// if configPath not defined, app will try use the value of $KUBECONFIG env
func ConnectToCluster(configPath string) (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	if configPath == "" {
		if os.Getenv("KUBECONFIG") == "" {
			kubePath := filepath.Join(fmt.Sprintf("%v/.kube/config", os.Getenv("HOME")))
			os.Setenv("KUBECONFIG", kubePath)
		}
		configPath = os.Getenv("KUBECONFIG")
	}
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return &kubernetes.Clientset{}, err
	}
	//creates clientSet
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &kubernetes.Clientset{}, err
	}
	return clientset, nil
}

func ErrHandler(err error) {
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}

func GetNodes(clientSet *kubernetes.Clientset) (*v1.NodeList, error) {
	nodes, err := clientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return &v1.NodeList{}, err
	}
	return nodes, nil
}

func GetPodNames(clientSet *kubernetes.Clientset, ns string) ([]string, error) {
	pods, err := clientSet.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []string{}, err
	}
	var podsNameList = make([]string, len(pods.Items))
	for counter, pod := range pods.Items {
		podsNameList[counter] = pod.Name
	}
	return podsNameList, nil
}

// No need to pass container name for single container pods, otherwise if not presented function will throw an error
func GetPodLogs(clientSet *kubernetes.Clientset, Ns string, podName string, containerName string) (string, error) {
	var req *rest.Request
	if containerName != "" {
		req = clientSet.CoreV1().Pods(Ns).GetLogs(podName, &v1.PodLogOptions{Container: containerName})
	} else {
		req = clientSet.CoreV1().Pods(Ns).GetLogs(podName, &v1.PodLogOptions{})
	}
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	// read the logs
	logs, err := io.ReadAll(podLogs)
	if err != nil {
		return "", err
	}
	return string(logs), nil
}

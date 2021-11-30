package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var waitGroup sync.WaitGroup
	var workNum = 1
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigPath())
	if err != nil {
		panic(err.Error())
	}

	// create the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	list(clientSet, &waitGroup, workNum)
	watch(clientSet, &waitGroup, workNum)

	waitGroup.Wait()
}

func list(clientSet *kubernetes.Clientset, wg *sync.WaitGroup, worker int) {
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			for {
				pods, err := clientSet.CoreV1().Pods("test").List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					panic(err.Error())
				}
				fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

				// Examples for error handling:
				// - Use helper functions like e.g. errors.IsNotFound()
				// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
				namespace := "test"
				n := rand.Intn(100)
				pod := "test-" + strconv.Itoa(n)
				_, err = clientSet.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
				} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
					fmt.Printf("Error getting pod %s in namespace %s: %v\n",
						pod, namespace, statusError.ErrStatus.Message)
				} else if err != nil {
					panic(err.Error())
				} else {
					fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
				}

				time.Sleep(2 * time.Second)
			}
		}()
	}
}

func watch(clientSet *kubernetes.Clientset, wg *sync.WaitGroup, worker int) {
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			watch, err := clientSet.CoreV1().Pods("test").Watch(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}
			for {
				select {
				case e := <-watch.ResultChan():
					fmt.Println(e)
				}
			}
		}()
	}
}

func getKubeConfigPath() string {
	return os.Getenv("KUBECONFIG")
}

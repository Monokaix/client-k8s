package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"client-k8s/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	workNum      = 150
	watchWorkNum = 150
	qps          = 100
	burst        = 100
	clientCount  = 150
)

func main() {
	flag.IntVar(&workNum, "worker", workNum, "number of current list workers")
	flag.IntVar(&watchWorkNum, "watchWorker", watchWorkNum, "number of current watch workers")
	flag.IntVar(&qps, "qps", qps, "kube client qps")
	flag.IntVar(&burst, "burst", burst, "kube client burst")
	flag.IntVar(&clientCount, "client-count", clientCount, "kube client count")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())
	var waitGroup sync.WaitGroup
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigPath())
	if err != nil {
		panic(err.Error())
	}
	config.QPS = float32(qps)
	config.Burst = burst
	pkg.MakeClient(clientCount, config)
	list(&waitGroup, workNum)
	watch(&waitGroup, watchWorkNum)
	waitGroup.Wait()
}

func list(wg *sync.WaitGroup, worker int) {
	for i := 0; i < worker; i++ {
		wg.Add(1)
		clientSet := pkg.GetClient(i, clientCount)
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
				//namespace := "test"
				//n := rand.Intn(100)
				//pod := "test-" + strconv.Itoa(n)
				//_, err = clientSet.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
				//if errors.IsNotFound(err) {
				//	fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
				//} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				//	fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				//		pod, namespace, statusError.ErrStatus.Message)
				//} else if err != nil {
				//	panic(err.Error())
				//} else {
				//	fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
				//}
				//
				// time.Sleep(time.Second)
			}
		}()
	}
}

func watch(wg *sync.WaitGroup, worker int) {
	for i := 0; i < worker; i++ {
		wg.Add(1)
		clientSet := pkg.GetClient(i, clientCount)
		go func() {
			_, err := clientSet.CoreV1().Pods("test").Watch(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}
			//select {
			//case _ = <-watch.ResultChan():
			//	fmt.Println("got one event")
			//}
		}()
	}
}

func getKubeConfigPath() string {
	return os.Getenv("KUBECONFIG")
}

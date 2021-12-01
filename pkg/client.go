package pkg

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var ClientSets []*kubernetes.Clientset

func MakeClient(count int, config *rest.Config) {
	ClientSets = make([]*kubernetes.Clientset, count)
	fmt.Println(config.QPS, config.Burst)
	for i := 0; i < count; i++ {
		client, err := kubernetes.NewForConfig(config)
		fmt.Printf("%p", client)
		if err != nil {
			panic(err)
		}
		ClientSets[i] = client
	}
}

func GetClient(i, count int) *kubernetes.Clientset {
	return ClientSets[i%count]
}

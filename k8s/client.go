package k8s

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func ReadK8s() {
	c := make(chan struct{})
	wait.Until(func() {
		fmt.Println("######################")
	}, time.Second, c)

}

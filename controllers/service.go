package controllers

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	queue "k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ServiceController struct {
	queue       queue.RateLimitingInterface
	syncService func(ctrl.Request) (ctrl.Result, error)
}

// func newServiceController() *ServiceController {
// 	return &ServiceController{
// 		queue: queue.NewNamedRateLimitingQueue(queue.DefaultControllerRateLimiter(), "service"),
// 	}
// }

func (c *ServiceController) add(obj interface{}) {
	// req ctrl.Request
	// fmt.Println("xxxx-add:", obj)
	if re, ok := obj.(ctrl.Request); ok {
		c.queue.Add(re)
	}
}

func (c *ServiceController) Run(workers int, stopCh <-chan struct{}) {
	// defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	// c.log.Info("Starting Service controller")
	// defer c.log.Info("Shutting down Service controller")

	// if !cache.WaitForNamedCacheSync("Service", stopCh, c.listerSynced) {
	// 	return
	// }

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *ServiceController) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *ServiceController) processNextWorkItem() bool {
	log := log.WithField("function", "ServiceController")

	key, quit := c.queue.Get()
	// fmt.Println("xxxx-processNextWorkItem:", key, quit)

	if quit {
		return false
	}
	defer c.queue.Done(key)
	if re, ok := key.(ctrl.Request); ok {
		c.syncService(re)
	}

	c.queue.Forget(key)
	log.Info("Successfully synced")

	return true
}

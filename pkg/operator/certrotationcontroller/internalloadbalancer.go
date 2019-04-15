package certrotationcontroller

import (
	"fmt"
	"strings"

	"k8s.io/klog"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

func (c *CertRotationController) syncInternalLoadBalancerHostnames() error {
	infrastructureConfig, err := c.infrastructureLister.Get("cluster")
	if err != nil {
		return err
	}
	hostname := infrastructureConfig.Status.APIServerURL
	hostname = strings.Replace(hostname, "https://", "", 1)
	hostname = hostname[0:strings.LastIndex(hostname, ":")]
	hostname = strings.Replace(hostname, "api.", "api-int.", 1)

	klog.V(2).Infof("syncing internal loadbalancer hostnames: %v", hostname)
	c.internalLoadBalancer.setHostnames([]string{hostname})
	return nil
}

func (c *CertRotationController) runInternalLoadBalancerHostnames() {
	for c.processExternalLoadBalancerHostnames() {
	}
}

func (c *CertRotationController) processInternalLoadBalancerHostnames() bool {
	dsKey, quit := c.internalLoadBalancerHostnamesQueue.Get()
	if quit {
		return false
	}
	defer c.internalLoadBalancerHostnamesQueue.Done(dsKey)

	err := c.syncInternalLoadBalancerHostnames()
	if err == nil {
		c.internalLoadBalancerHostnamesQueue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.internalLoadBalancerHostnamesQueue.AddRateLimited(dsKey)

	return true
}

// eventHandler queues the operator to check spec and status
func (c *CertRotationController) internalLoadBalancerHostnameEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.internalLoadBalancerHostnamesQueue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.internalLoadBalancerHostnamesQueue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.internalLoadBalancerHostnamesQueue.Add(workQueueKey) },
	}
}

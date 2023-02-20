package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	v1 "operator-crd/pkg/apis/example.com/v1"
	"time"

	informers "operator-crd/pkg/client/informers/externalversions/example.com/v1"
)

type Controller struct {
	informer  informers.BarInformer
	workqueue workqueue.RateLimitingInterface
}

func NewController(informer informers.BarInformer) *Controller {
	controller := &Controller{
		informer: informer,
		// WorkQueue 的实现，负责同步 Informer 和控制循环之间的数据
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "bar"),
	}

	klog.Info("Setting up Bar event handlers")

	// informer 注册了三个 Handler（AddFunc、UpdateFunc 和 DeleteFunc）
	// 分别对应 API 对象的“添加”“更新”和“删除”事件。
	// 而具体的处理操作，都是将该事件对应的 API 对象加入到工作队列中
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addBar,
		UpdateFunc: controller.updateBar,
		DeleteFunc: controller.deleteBar,
	})
	return controller
}

func (c *Controller) Run(thread int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShuttingDown()

	// 记录开始日志
	klog.Info("Starting Bar control loop")
	klog.Info("Waiting for informer caches to sync")

	// 等待缓存同步数据
	if ok := cache.WaitForCacheSync(stopCh, c.informer.Informer().HasSynced); !ok {
		return fmt.Errorf("failed to wati for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < thread; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

// runWorker 是一个不断运行的方法，并且一直会调用 c.processNextWorkItem 从 workqueue读取消息
func (c *Controller) runWorker() {
	for c.processNExtWorkItem() {
	}
}

// 从workqueue读取和读取消息
func (c *Controller) processNExtWorkItem() bool {
	// 获取 item
	item, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	if err := func(item interface{}) error {
		// 标记以及处理
		defer c.workqueue.Done(item)
		var key string
		var ok bool
		if key, ok = item.(string); !ok {
			// 判读key的类型不是字符串，则直接丢弃
			c.workqueue.Forget(item)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", item))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s':%s", item, err.Error())
		}
		c.workqueue.Forget(item)
		return nil
	}(item); err != nil {
		runtime.HandleError(err)
		return false
	}
	return true
}

// 尝试从 Informer 维护的缓存中拿到了它所对应的 Bar 对象
func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid respirce key:%s", key))
		return err
	}

	bar, err := c.informer.Lister().Bars(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// 说明是在删除事件中添加进来的
			return nil
		}
		runtime.HandleError(fmt.Errorf("failed to get bar by: %s/%s", namespace, name))
		return err
	}
	fmt.Printf("[BarCRD] try to process bar:%#v ...", bar)
	// 可以根据bar来做其他的事。
	// todo
	return nil
}

func (c *Controller) addBar(item interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(item); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) deleteBar(item interface{}) {
	var key string
	var err error
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(item); err != nil {
		runtime.HandleError(err)
		return
	}
	fmt.Println("delete crd")
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) updateBar(old, new interface{}) {
	oldItem := old.(*v1.Bar)
	newItem := new.(*v1.Bar)
	// 比较两个资源版本，如果相同，则不处理
	if oldItem.ResourceVersion == newItem.ResourceVersion {
		return
	}
	c.workqueue.AddRateLimited(new)
}

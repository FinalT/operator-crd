package main

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	clientset "operator-crd/pkg/client/clientset/versioned"
	"operator-crd/pkg/client/informers/externalversions"
	"time"

	"os"
	"os/signal"
	"syscall"
)

var (
	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

// 注册 SIGTERM 和 SIGINT 信号
// 返回一个 stop channel， 该通道在捕获到第一个信号时被关闭
// 如果捕获到第二个信号，程序直接退出
func setupSignalHandler() (stopCh <-chan struct{}) {
	// 当调用两次的时候 panics
	close(onlyOneSignalHandler)

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)

	// Notify 函数让 signal 包将输入信号转发到c
	// 如果没有列出要传递的信号，会将所有输入信号传递到 c； 否则只会传递列出的输入信号
	signal.Notify(c, shutdownSignals...)

	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // 第二个信号直接退出
	}()

	return stop
}

func main() {
	stopCh := setupSignalHandler()

	// 获取config
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		klog.Fatalln(err)
	}

	// 通过config构建clientSet
	// 这里的clientSet 是 Bar 的
	clientSet, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Fatalln(err)
	}

	// informerFactory 工厂类， 这里注入我们通过代码生成的 client
	// client 主要用于和 API Server 进行通信，实现 ListAndWatch
	factory := externalversions.NewSharedInformerFactory(clientSet, time.Second*30)

	// 实例化自定义控制器
	controller := NewController(factory.Example().V1().Bars())

	// 启动 informer，开始list 和 watch
	go factory.Start(stopCh)

	// 启动控制器
	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

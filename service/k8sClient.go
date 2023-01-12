package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wonderivan/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"kubez-backend/config"
)

type k8s struct {
	// 提供多集群client
	ClientMap map[string]*kubernetes.Clientset
	// 提供集群列表功能
	KubeConfMap map[string]string
}

var K8s k8s

// 根据集群名获取client
func (k *k8s) GetClient(cluster string) (*kubernetes.Clientset, error) {
	client, ok := k.ClientMap[cluster]
	if !ok {
		return nil, errors.New(fmt.Sprintf("集群: %s不存在, 无法获取client\n", cluster))
	}
	return client, nil
}

// 初始化client
func (k *k8s) Init() {
	mp := make(map[string]string, 0)
	k.ClientMap = make(map[string]*kubernetes.Clientset, 0)
	// 反序列化
	if err := json.Unmarshal([]byte(config.Kubeconfigs), &mp); err != nil {
		panic("kubeconfig反序列化失败")
	}
	k.KubeConfMap = mp
	// 初始化client
	for cluster, path := range mp {
		config, err := clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			panic(fmt.Sprintf("集群%s: 创建config失败 %s\n", cluster, err))
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(fmt.Sprintf("集群%s: 创建clientset失败 %s\n", cluster, err))
		}
		k.ClientMap[cluster] = clientset
		logger.Info(fmt.Sprintf("集群%s: 创建clientset成功", cluster))
	}

}

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wonderivan/logger"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kubez-backend/config"
)

type pod struct{}

type PodsResp struct {
	Items []corev1.Pod `json:"items"`
	Total int
}

var Pod pod

// 定义DataCell到Pod类型转换的方法
func (p *pod) toCells(std []corev1.Pod) []DataCell {
	cells := make([]DataCell, len(std))
	for i := range std {
		cells[i] = podCell(std[i])
	}
	return cells
}

func (p *pod) fromCells(cells []DataCell) []corev1.Pod {
	pods := make([]corev1.Pod, len(cells))
	for i := range cells {
		pods[i] = corev1.Pod(cells[i].(podCell))
	}
	return pods
}

// GetPods 获取pod列表，client 用于选择哪个集群
func (p *pod) GetPods(client *kubernetes.Clientset, filterName, namespace string, limit, page int) (podsResp *PodsResp, err error) {
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Info(fmt.Sprintf("获取Pod列表失败: %s\n", err))
		return nil, errors.New(fmt.Sprintf("获取Pod列表失败: %s\n", err))
	}

	// 实例化dataSelector对象
	selectableData := &dataSelector{
		GenericDataList: p.toCells(podList.Items),
		dataSelectorQuery: &DataSelectorQuery{
			FilterQuery: &FilterQuery{filterName},
			PaginateQuery: &PaginateQuery{
				limit,
				page,
			},
		},
	}
	// 先过滤
	filtered := selectableData.Filter()
	total := len(filtered.GenericDataList)
	// 再去排序和分页
	data := filtered.Sort()
	// 将[]DataCell类型的pod列表转为v1.pod列表
	pods := p.fromCells(data.GenericDataList)

	return &PodsResp{
		pods,
		total,
	}, nil

}

// GetPodDetail 获取pod详情
func (p *pod) GetPodDetail(client *kubernetes.Clientset, podName, namespace string) (*corev1.Pod, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		logger.Info(fmt.Sprintf("获取pod详情失败: %s\n", err))
		return nil, errors.New(fmt.Sprintf("获取pod详情失败: %s\n", err))
	}
	return pod, nil
}

// DeletePod 删除pod
func (p *pod) DeletePod(client *kubernetes.Clientset, podName, namespace string) error {
	if err := client.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{}); err != nil {
		logger.Info(fmt.Sprintf("删除pod: %s失败\n", err))
		return errors.New(fmt.Sprintf("删除pod: %s失败\n", err))
	}
	return nil
}

// UpdatePod 更新pod, content就是pod的整个json体
func (p *pod) UpdatePod(client *kubernetes.Clientset, podName, namespace, content string) error {
	pod := &corev1.Pod{}
	// 发序列化
	if err := json.Unmarshal([]byte(content), pod); err != nil {
		logger.Info("pod json反序列化失败: %s\n", err)
		return errors.New(fmt.Sprintf("pod json反序列化失败: %s\n", err))
	}
	// 更新pod
	_, err := client.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		logger.Info(fmt.Sprintf("更新pod失败: %s\n", err))
		return errors.New(fmt.Sprintf("更新pod失败: %s\n", err))
	}
	return nil
}

// GetContainers 获取pod中的容器名
func (p *pod) GetContainers(client *kubernetes.Clientset, podName, namespace string) (containers []string, err error) {
	// 获取pod对象
	pod, err := p.GetPodDetail(client, podName, namespace)
	if err != nil {
		return nil, err
	}
	// 从pod对象中获取container
	Containers := pod.Spec.Containers
	for _, container := range Containers {
		containers = append(containers, container.Name)
	}
	return containers, nil
}

// GetPodLog 获取pod中的容器日志
func (p *pod) GetPodLog(client *kubernetes.Clientset, containerName, podName, namespace string) (log string, err error) {
	//设置日志的配置，容器名以及tail的行数
	LineLimit := int64(config.PodLogTailLine)
	option := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &LineLimit,
	}
	// 获取request实例
	req := client.CoreV1().Pods(namespace).GetLogs(containerName, option)
	// 发起request请求，返回一个ioReadCloser类型（等同于response.body）
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		logger.Error(fmt.Sprintf("获取PodLog失败, %s\n", err))
		return "", errors.New(fmt.Sprintf("获取PodLog失败, %s\n", err))
	}
	defer podLogs.Close()
	// 将response body写入缓冲区，目的是为了转成string返回
	buf := new(bytes.Buffer)
	io.Copy(buf, podLogs)
	return buf.String(), nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/wonderivan/logger"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type deployment struct{}

type DepResp struct {
	Items []appv1.Deployment `json:"items"`
	Total int
}

var Deployment deployment

// 定义DataCell到Deployment类型转换的方法
func (dep *deployment) toCells(deps []appv1.Deployment) []DataCell {
	cells := make([]DataCell, len(deps))
	for i := range deps {
		cells[i] = deploymentCell(deps[i])
	}
	return cells
}

func (dep *deployment) fromCells(cells []DataCell) []appv1.Deployment {
	deps := make([]appv1.Deployment, len(cells))
	for i := range cells {
		deps[i] = appv1.Deployment(cells[i].(deploymentCell))
	}
	return deps
}

// 获取deployment列表
func (dep *deployment) GetDeployment(client *kubernetes.Clientset, filterName, namespace string, limit, page int) (depResp *DepResp, err error) {
	depList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Info(fmt.Sprintf("获取deployment列表失败: %s\n", err))
		return nil, errors.New(fmt.Sprintf("获取deployment列表失败: %s\n", err))
	}
	// 实例化dataSelector对象
	selectableData := &dataSelector{
		GenericDataList: dep.toCells(depList.Items),
		dataSelectorQuery: &DataSelectorQuery{
			FilterQuery: &FilterQuery{Name: filterName},
			PaginateQuery: &PaginateQuery{
				limit,
				page,
			},
		},
	}
	// 先过滤
	filtered := selectableData.Filter()
	total := len(filtered.GenericDataList)
	// 在排序和分页
	data := filtered.Sort()
	// 将[]DataCell类型的pod列表转为v1.pod列表
	deployments := dep.fromCells(data.GenericDataList)

	return &DepResp{
		Items: deployments,
		Total: total,
	}, nil
}

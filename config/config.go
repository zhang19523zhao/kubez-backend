package config

const (
	ListenAddr  = "localhost:9090"
	Kubeconfigs = `{"Kube1": "/Users/zhanghao/.kube/config", "Kube2": "/Users/zhanghao/.kube/config"}`
	// 查看日志时显示的行数 tail -n
	PodLogTailLine = 5000
)

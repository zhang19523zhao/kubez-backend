package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"kubez-backend/config"
	"kubez-backend/controller"
	"kubez-backend/service"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// 初始化gin 对象
	r := gin.Default()
	// 初始化k8 sclientset
	service.K8s.Init()
	// 初始化路由
	controller.Router.InintApiRouter(r)
	// 启动gin server
	srv := &http.Server{
		Addr:    config.ListenAddr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen: ", err)
		}
	}()
	// 优雅关闭
	// 声明一个系统信号的channel，并监听他，如果没有信号就一直阻塞，如果有就继续执行
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	// 设置ctx超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// cancel 用于释放ctx
	defer cancel()
	// 关闭gin
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Gin Server 关闭异常: ", err)
	}
	logger.Info("Gin Server退出成功")
}

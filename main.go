package main

import (
	"fmt"
	"web/controller"
	"web/dao/mysql"
	"web/dao/redis"
	"web/logger"
	"web/pkg/snowflake"
	"web/routes"
	"web/settings"

	"go.uber.org/zap"
)

//go web开发通用的脚手架模板

func main() {
	//1.加载日志
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed,err:%v\n", err)
		return
	}
	//2.初始化日志
	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("logger init failed,err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success...")

	//3.初始化MySQL连接
	if err := mysql.Init(); err != nil {
		fmt.Printf("init mysql failed,err:%v\n", err)
		return
	}
	defer mysql.Close()

	//4.初始化Redis连接
	if err := redis.Init(); err != nil {
		fmt.Printf("init redis failed,err:%v\n", err)
		return
	}
	defer redis.Close()

	if err := snowflake.Init(settings.Conf.StartTime, settings.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed,err:%v\n", err)
		return
	}
	//初始化gin框架内置的校验器使用的翻译器
	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init validator trans failed,err:%v\n", err)
		return
	}
	//5.注册路由
	r := routes.SetupRouter(settings.Conf.Mode)
	err := r.Run(fmt.Sprintf(":%d", settings.Conf.AppConfig.Port))
	if err != nil {
		fmt.Printf("run server failed,err:#{err}\n")
		return
	}
}

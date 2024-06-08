package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// 定义一个全局对象db
var db *sqlx.DB

// 定义一个初始化数据库的函数
func Init() (err error) {
	// dsn:定义的数据库名称
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&parseTime=True",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.dbname"),
	)
	//也可以使用MustConnect连接不成功就panic
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		zap.L().Error("connect DB failed", zap.Error(err))
		fmt.Printf("connect to DB failed,err:%v\n", err)
		return
	}
	db.SetMaxOpenConns(viper.GetInt("mysql.max_open_conns"))
	db.SetMaxIdleConns(viper.GetInt("mysql.max_idle_conns"))
	return nil
}

func Close() {
	_ = db.Close()
}

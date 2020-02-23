package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/read-loadbalance/lb"
	"log"
)

var slaveLB *lb.SlaveLoadBalancer

func main() {
	//构造数据源
	creator := func() *sql.DB {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&autocommit=1", "root", 123456, "localhost", 3306, "test", "utf8mb4")
		conn, er := sql.Open("mysql", dsn)
		if er != nil {
			log.Fatalf("create db conn error:%v", er.Error())
		}
		return conn
	}

	slave1 := creator()
	slave2 := creator()

	//jumpB数仓拉取的开始时间
	//jumpE数仓拉取的结束时间
	//BigDaPullInd哪一个从库是用来拉取的
	slaveLB = lb.BuildSlaveLoadBalancer(1, 2, 1, slave1, slave2)

	//获取数据源
	slaveLB.GetPollingNode()

}

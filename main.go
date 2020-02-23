package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
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

	//构造slave-lb
	slaveLB = lb.BuildSlaveLoadBalancer(1, 2, 1, slave1, slave2)

	//获取数据源
	_, _, conn := slaveLB.GetPollingNode()

	//放入任意orm中，这里举例dbr
	orm := &dbr.Connection{DB: conn}
	orm.NewSession(nil).SelectBySql("select id from tb_orders limit 1")

}

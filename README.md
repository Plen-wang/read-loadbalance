# read-loadbalance
master-slave(n) 读库集群负载均衡器，使用简单轮询。

# 使用场景
> 1.一般我们会有多个从库，需要在从库的读取上做负载均衡。  
> 2.在数仓拉取数据的时候经常对产线DB造成影响，所以会独立一个从库专门用来拉取，但是这个从库的利用率非常低。
数仓拉取数据一般在业务低峰期进行，iops峰值较高，但是持续时间很短。我们可以错开这个时间段，让这台从库的利用率最大化。
![vim](https://raw.githubusercontent.com/Plen-wang/blogsImage/master/githubimages/lb/lb.png)


```
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

```
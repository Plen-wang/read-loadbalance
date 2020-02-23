# read-loadbalance
master-slave(n) 读库集群负载均衡器，使用简单轮询。

# 使用场景
> 1.一般我们会有多个从库，需要在从库的读取上做负载均衡。  
> 2.在数仓拉取数据的时候经常对产线DB造成影响，所以会独立一个从库专门用来拉取，但是这个从库的利用率非常低。
数仓拉取数据一般在业务低峰期进行，iops峰值较高，但是持续时间很短。我们可以错开这个时间段，让这台从库的利用率最大化。

demo:
```
package main

import (
	"database/sql"
	"fmt"
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

```
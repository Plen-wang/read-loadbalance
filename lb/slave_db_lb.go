package lb

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

//slave 读节点负载均衡（简单轮询）。
//避开在某个时间段读取集群中某一个节点。 该节点正在被大数据拉取数据。
type SlaveLoadBalancer struct {
	//轮询index
	pollingIndex int
	//轮询lock
	pollingLock *sync.Mutex
	//slave cluster
	rdbCluster []*sql.DB
	//大数据拉取的节点index
	bigDataPullIndex int
	//跳过的开始时间。数仓开始拉取的时间。24h区间。
	jumpBegin int
	//跳过的结束时间。数仓技术拉取的时间。24h区间。
	jumpEnd int
}

//构造 SlaveLoadBalance 实例。
//jumpB、jumpE 开始时间和结束时间不支持跨天。
//根据经验数仓拉取数据基本都在0点到1点之间，或1点到2点之间，集中在凌晨。也可以自行调整。
func BuildSlaveLoadBalancer(jumpB, jumpE, BigDaPullInd int, slaveNodeConn ...*sql.DB) *SlaveLoadBalancer {
	CheckParam(jumpB, jumpE, BigDaPullInd, slaveNodeConn)

	sInstance := &SlaveLoadBalancer{}
	sInstance.pollingIndex = 0
	sInstance.pollingLock = &sync.Mutex{}
	//保证插入的顺序
	for i := 0; i < len(slaveNodeConn); i++ {
		sInstance.rdbCluster = append(sInstance.rdbCluster, slaveNodeConn[i])
	}
	sInstance.bigDataPullIndex = BigDaPullInd
	sInstance.jumpBegin = jumpB
	sInstance.jumpEnd = jumpE

	return sInstance
}

//循环获取下一个节点
func (t *SlaveLoadBalancer) GetPollingNode() (curIndex int, hit bool, conn *sql.DB) {
	defer t.pollingLock.Unlock()
	t.pollingLock.Lock()

	//返回当前index节点
	//如果一圈还没轮询完，继续向前轮询
	polling := func() {
		curIndex = t.pollingIndex
		conn = t.rdbCluster[t.pollingIndex]
		if t.pollingIndex+1 < len(t.rdbCluster) {
			t.pollingIndex++
		} else {
			//一圈结束，重新开始
			t.pollingIndex = 0
		}
	}
	polling()
	//是否命中拉取节点，如果命中跳过当前节点。
	hit = t.HitJumpTimeRange(curIndex)
	if hit {
		polling()
	}

	return
}

//是否命中数仓拉取的时间段和节点索引
func (t *SlaveLoadBalancer) HitJumpTimeRange(curIndex int) (hit bool) {
	h := time.Now().Hour()
	//是否命中数据拉取节点&&是否在拉取时间段内
	if t.bigDataPullIndex == curIndex && t.jumpBegin == h && t.jumpEnd > h {
		hit = true
	}
	return
}

//check param
func CheckParam(jumpB int, jumpE int, BigDaPullInd int, slaveNodeConn []*sql.DB) {
	if jumpB < 0 || jumpB > 23 {
		panic(fmt.Sprintf("%v.build SlaveLoadBalance error.\"jumpB\" param invalid.", jumpB))
	}
	if jumpE < 0 || jumpE > 23 {
		panic(fmt.Sprintf("%v.build SlaveLoadBalance error.\"jumpE\" param invalid.", jumpE))
	}
	if jumpE < jumpB {
		panic(fmt.Sprintf("jumpB:%v,jumpE:%v.build SlaveLoadBalance error.\"jumpE<jumpB\" param invalid.", jumpB, jumpE))
	}
	if BigDaPullInd < 0 {
		panic(fmt.Sprintf("%v.build SlaveLoadBalance error.\"bDataPullInd\" param invalid.", BigDaPullInd))
	}
	if slaveNodeConn == nil || len(slaveNodeConn) <= 0 {
		panic("build SlaveLoadBalance error.\"slaveNodeConn\" param is nil.")
	}
}




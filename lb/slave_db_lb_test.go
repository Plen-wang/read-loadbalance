package lb

import (
	"database/sql"
	"reflect"
	"testing"
)

//concurrency test
func BenchmarkSlaveLoadBalancer_GetPollingNode(t *testing.B) {
	dbr0 := &sql.DB{}
	dbr1 := &sql.DB{}
	dbr2 := &sql.DB{}

	slb := BuildSlaveLoadBalancer(15, 16, 1, dbr0, dbr1, dbr2)

	i0 := 0
	i1 := 0
	i2 := 0
	i3 := 0
	hit := 0
	for i := 0; i < t.N; i++ {
		lbIndex, h, _ := slb.GetPollingNode()
		if lbIndex == 0 {
			i0++
		} else if lbIndex == 1 {
			i1++
		} else if lbIndex == 2 {
			i2++
		} else if lbIndex == 3 {
			i3++
		}
		if h {
			hit++
		}
	}

	t.Logf("0:%v,1:%v,2:%v,3:%v,hit:%v", i0, i1, i2, i3, hit)
}

//test connection append 顺序
func TestSlaveLoadBalancer_GetPollingNode_append_order(t *testing.T) {
	dbr0 := &sql.DB{}
	dbr0.SetMaxIdleConns(1)
	dbr1 := &sql.DB{}
	dbr1.SetMaxIdleConns(2)
	dbr2 := &sql.DB{}
	dbr2.SetMaxIdleConns(3)

	slb := BuildSlaveLoadBalancer(15, 16, 1, dbr0, dbr1, dbr2)

	for i := 0; i < len(slb.rdbCluster); i++ {
		v := reflect.ValueOf(slb.rdbCluster[i])
		if v.Kind() == reflect.Ptr {
			maxV := v.Elem().FieldByName("maxIdle")
			if maxV.Int() == int64(i+1) {
				t.Log(maxV.Int())
			}
		}
	}
}

//test hit
func TestSlaveLoadBalance_JumpTimeRange(t *testing.T) {
	dbr0 := &sql.DB{}
	dbr1 := &sql.DB{}
	dbr2 := &sql.DB{}
	var slb = BuildSlaveLoadBalancer(15, 16, 1, dbr0, dbr1, dbr2)

	hit := slb.HitJumpTimeRange(1)

	t.Logf("hit:%v", hit)
}

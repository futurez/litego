package mysqlz

import (
	"sync"
	"testing"
)

func TestStoreProceduce(t *testing.T) {
	connpool, err := NewMysqlConnPool("fcloud", "fcloud2015", "192.168.1.141", "3306", "gotye_open_live", "utf8", 20)
	if err != nil {
		t.Error(err.Error())
		return
	}

	var wait sync.WaitGroup

	liveroomIds := [4]int64{210150, 210060, 210256, 210123}

	for i := 0; i <= 7; i++ {
		wait.Add(1)
		go func(n int) {
			defer wait.Done()
			t.Log("start n = ", n, "liveroomid = ", liveroomIds[n%4])

			model := connpool.GetModel()

			tx, err := model.GetDB().Begin()
			if err != nil {
				t.Error(err.Error())
				return
			}
			defer tx.Commit()

			_, err = tx.Exec("call follow_liveroom_num(?, @count)", liveroomIds[n%4])
			if err != nil {
				t.Error(err.Error())
				return
			}

			var count int
			err = tx.QueryRow("select @count").Scan(&count)
			if err != nil {
				t.Error(err.Error())
				return
			}
			t.Log("end n = ", n, "liveroomid = ", liveroomIds[n%4], "count =", count)
		}(i)
	}
	wait.Wait()
}

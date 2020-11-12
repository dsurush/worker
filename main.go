package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
	"log"
	"net/http"
	"sync/atomic"
	"time"
	_ "worker/loginit"
	"worker/models"
)

// Worker - воркер который запускает каждые d(duration) функцию f если он еще не запущен
func worker(d time.Duration, async bool, f func(pool *pgxpool.Pool), pool *pgxpool.Pool) {
	var reentranceFlag int64
	for range time.Tick(d) {
		go func() {
			if !async {

				if atomic.CompareAndSwapInt64(&reentranceFlag, 0, 1) {
					defer atomic.StoreInt64(&reentranceFlag, 0)
				} else {
					log.Println("Previous worker in process now")
					return
				}
			}
			f(pool)
		}()
	}
}

func GetActivitiesFromDB(pool *pgxpool.Pool) (Activities []models.Activity, err error) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Printf("can't get connection %e", err)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), `select *from activities where exited = false`)
	if err != nil {
		fmt.Printf("can't read user rows %e\n", err)
		log.Printf("can't read user rows %e\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		Activity := models.Activity{}
		err := rows.Scan(
			&Activity.ID,
			&Activity.UserId,
			&Activity.Token,
			&Activity.UnixTime,
			&Activity.Status,
			&Activity.WorkTime,
			&Activity.Exited,
		)
		if err != nil {
			fmt.Println("can't scan err is = ", err)
			log.Println("can't scan err is = ", err)
		}
		Activities = append(Activities, Activity)
	}
	if rows.Err() != nil {
		fmt.Printf("rows err %s\n", err)
		log.Printf("rows err %s\n", err)
		return nil, rows.Err()
	}
	return
}

func SendRequests(pool *pgxpool.Pool) {
	//	 []models.Activity
	Activities, err := GetActivitiesFromDB(pool)
	//fmt.Println(Activities)
	if err != nil {
		fmt.Println(`Can't get from db data' `, err)
		log.Println(`Can't get from db data' `, err)
		return
	}
	for _, value := range Activities {
		if time.Now().Unix()-value.UnixTime > 130 {
			ok := DoRequest(value.WorkTime, value.Status, value.Token, value.UserId)
			if !ok {
				fmt.Println(`Cant send request to user with user_id = `, value.UserId)
				log.Println(`Cant send request to user with user_id = `, value.UserId)
			}
		}
	}
	return
}

func DoRequest(time int64, status bool, token string, userId int64) bool {
	client := &http.Client{}
	body := fmt.Sprintf(`{
"time" : %d,
"status" : %t
}`, time, status)
	AuthToken := fmt.Sprintf(`Bearer %s`, token)
	//		body := `{
	// "time" : 360,
	// "status" : false
	//}`
	req, err := http.NewRequest(
		"POST", "http://127.0.0.1:8888/api/exit", bytes.NewBuffer([]byte(body)),
	)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", AuthToken)
	//	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6OCwiZXhwIjoxNjAyNjM5ODA3LCJsb2dpbiI6InRlc3QiLCJyb2xlIjoidXNlciJ9.HQtfZ1bqtEw-JR4YmAlJRGoHyTxRUCeNrAMIbSqTvfg")
	if err != nil {
		fmt.Printf("Can't build request\n")
		log.Printf("Can't build request\n")
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Can't send Request\n")
		return false
	}
	fmt.Println(`Status Code = `, resp.StatusCode, ` UserId = `, userId)
	log.Println(`Status Code = `, resp.StatusCode, ` UserId = `, userId)

	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}

func Start(pool *pgxpool.Pool) {
	worker(15*time.Second, true, SendRequests, pool)
}

func main() {

	pool, err := pgxpool.Connect(context.Background(), `postgres://dsurush:dsurush@localhost:5432/ccs?sslmode=disable`)
	if err != nil {
		log.Printf("Owibka - %e", err)
		log.Fatal("Can't Connection to DB")
	} else {
		log.Println(`CONNECTION TO DB IS SUCCESS`)
		fmt.Println("CONNECTION TO DB IS SUCCESS")
	}
	log.Println(`Job is started`)
	Start(pool)
}

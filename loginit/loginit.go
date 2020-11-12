package loginit

import (
	"github.com/natefinch/lumberjack"
	_ "github.com/natefinch/lumberjack"
	"log"
)

func init(){
	//	fmt.Println("DO")
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/home/ccs/exitjob/logs/foo.log",
//		Filename:   "C:/Users/Professional/go/src/worker/foo.log",
		MaxSize:    20, // megabytes
		MaxBackups: 5,
		MaxAge:     60, //days
		Compress:   true, // disabled by default
	})
	//	fmt.Println("DID")
}
package mq

import (
	"log"
)

var (
	done chan bool
)

// StartConsume 文件异步传输main函数调用，表示传输开始消费
func StartConsume(qName, cName string, callback func(msg []byte)bool)  {
	//1. 通过channel.consume获得消息信道
	// consume 返回一个 <-chan， error
	msgs, err := channel.Consume(
		qName,
		cName,
		true,
		false,
		false,
		false,
		nil,
		)
	if err != nil {

		log.Fatal(err.Error())
		return
	}
	//2. 循环获取队列的消息
	done = make(chan bool)
	go func() {
		// 没有消息的时候会阻塞该goroutine
		for msg := range msgs{
			//3. 调用callback
			processSuc := callback(msg.Body)
			if !processSuc{
				//TODO:将任务写到另一个队列，用于异常情况的重试
			}
		}
	}()

	<-done
	channel.Close()
}

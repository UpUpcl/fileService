package mq

import (
	"filestore-server/config"
	"github.com/streadway/amqp"
	"log"
)

var conn *amqp.Connection
var channel *amqp.Channel
var notifyClose chan *amqp.Error

func init()  {
	if !config.AsyncTransferEnable{
		return
	}
	if initChannel(){
		channel.NotifyClose(notifyClose)
	}

	go func() {
		for{
			select {
			case msg := <- notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed:%+v\n", msg)
				initChannel()

			}
		}
	}()
}

func initChannel() bool {
	//1. 判断channel是否过
	if channel !=nil{
		return true
	}
	//2. 获得rabbitmq的一个连接
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//3. 打开一个channel，用于消息的发布与接受等
	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// Publish 发布消息
func Publish(exchange, routingKey string, msg []byte) bool {
	//1. 判断channel是否正常
	if !initChannel(){
		return false
	}
	//2. 执行消息发布动作
	err := channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType:     "text/plain",
		Body:            msg,
	})
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}


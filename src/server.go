package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) //链接客户
var broadcast = make(chan Message) //广播消息

//配置
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool{
		return true
	},
}

//定义信息结构
type Message struct {
	Username string `json:"username"`
	Message  string  `json:"message"`
}

func main(){
	//创建服务器文件夹
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/",fs)

	//配置websocket路由
	http.HandleFunc("/ws",handleConnections)

	//开始监听传入的信息
	go handleMessages()

	//监听端口启动服务器错误记录
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000",nil)
	if err != nil{
		log.Fatal("ListenAndServer:", err)
	}
}
//注册websocket
func handleConnections(w http.ResponseWriter,r *http.Request){
	ws ,err := upgrader.Upgrade(w,r,nil)
	if err != nil{
		log.Fatal(err)
	}
	//确保在函数返回时关闭连接
	defer ws.Close()

	//注册新客户
	clients[ws] = true
	//不断的从页面上获取数据 然后广播发送出去
	for {
		var msg Message
		//以JSON格式读取新消息并将其映射到Message对象
		err := ws.ReadJSON(&msg)
		if err != nil{
			log.Printf("error:%v",err)
			delete(clients,ws)
			break
		}
		//新接收的消息发送到广播频道
		broadcast <- msg

	}
}

func handleMessages(){
	for {
		//从广播获取下一条消息
		msg := <- broadcast
		//将其发送给当前连接的每个客户端
		for client := range clients{
			err := client.WriteJSON(msg) //核心 发送
			if err != nil{
				log.Printf("error:%v",err)
				client.Close()
				delete(clients,client)
			}
		}
	}
}





























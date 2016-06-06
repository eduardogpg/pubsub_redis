package main

import (
	"log"
	"gopkg.in/redis.v3"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

type Cliente struct {
	Id int
	websocket *websocket.Conn
}
var Clientes = make(map[int]Cliente)


type Request struct{
	Id 		int 		`json:"id"`
	Name 	string 	`json:"name"`
}

func ConnectNewClient(request_chanel chan Request){
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	
	pubsub, err := client.Subscribe("real_time_example")
	if err != nil {
			log.Println("problem redis server")
			panic(err)
	}

	for{
		message, err := pubsub.ReceiveMessage()
		if err != nil{
			log.Println("problem redis server")
			panic(err)
		}
		log.Println(message.Channel)
		log.Println(message.Pattern)
		log.Println(message.Payload)

		request := Request{}
		if err := json.Unmarshal([]byte(message.Payload), &request); err != nil {
			log.Println("no es posible realizar el unmarshal")
			panic(err)
		}
		log.Println(request.Name)
		request_chanel <- request
	}
}

func main() {
	channel_request := make(chan Request)
	go ConnectNewClient(channel_request)
	go ValidateChannel(channel_request)

	mux := mux.NewRouter()
	mux.HandleFunc("/subscribe/", Subscribe).Methods("GET")
	http.Handle("/", mux)
	log.Println("El servidor se encuentra a la escucha en el puerto 10443")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func ValidateChannel(request chan Request){
	for{
		select {
			case r := <- request:
				SendMessage(r)
		}
	}
}

func SendMessage(request Request){
	for _, cliente := range Clientes {
		if err := cliente.websocket.WriteJSON(request); err != nil {
			return
		}
	}	
}

func Subscribe(w http.ResponseWriter, r *http.Request){
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		return
	}
	
	log.Println("La parte magica")

	cliente := Cliente{ 0, ws }
	Clientes[0] = cliente
 	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(Clientes, cliente.Id)
			return
		}
	}
}

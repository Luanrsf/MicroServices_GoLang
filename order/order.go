package main

import (
	"io/ioutil"
	"fmt"
	"net/http"
	"os"
	"github.com/streadway/amqp"
	"time"
	"encoding/json"
	"github.com/nu7hatch/gouuid"
	"github.com/go-redis/redis/v7"
	
)

type Product struct {
	Uuid    string  `json:"uuid"`
	Product string  `json:"product"`
	Price   float32 `json:"price,string"`
}

type Order struct {
	Uuid string
	Name string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	ProductId string `json:"product_id"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"create_at,string"`

}
var productsUrl string
func init(){
	productsUrl = os.Getenv("PRODUCT_URL")
}
func DBConnect() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: "",
		DB:       0,
	})
	return client
}

func Connect() *amqp.Channel {
	dsn := "amqp://" + os.Getenv("RABBITMQ_DEFAULT_USER") + ":" + os.Getenv("RABBITMQ_DEFAULT_PASS") + "@" + os.Getenv("RABBITMQ_DEFAULT_HOST") + ":" + os.Getenv("RABBITMQ_DEFAULT_PORT") + os.Getenv("RABBITMQ_DEFAULT_VHOST")

	conn, err := amqp.Dial(dsn)

	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}

	channel, err := conn.Channel()
	if err != nil {
		panic(err.Error())
	}
	return channel
}

func Notify(payload []byte, exchange string, routingKey string, ch *amqp.Channel) {

	err := ch.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(payload),
		})

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Message sent")
}

func StartConsuming(ch *amqp.Channel,in chan []byte){
	q, err:= ch.QueueDeclare(
		os.Getenv("RABBITMQ_COMSUMER_QUEUE"),
		true,
		false,
		false,
		false,
		nil,
	)
		if err != nil {
		panic(err.Error())
	}
	msgs, err := ch.Consume(
		q.Name,
		"checkout",
		true,
		false,
		false,
		false,
		nil,
	)
		if err != nil {
		panic(err.Error())
	}
	go func ()  {
		for m:=range msgs{
			in<-[]byte(m.Body)
		}
		close(in)
	}()
	
}
func createOrder(payload []byte) Order {
	var order Order
	json.Unmarshal(payload, &order)

	uuid, _ := uuid.NewV4()
	order.Uuid = uuid.String()
	order.Status = "pendente"
	order.CreatedAt = time.Now()
	saveOrder(order)
	return order

}
func saveOrder(order Order){
	json,_ :=json.Marshal(order)
	connection:=DBConnect()
	err:=
	connection.Set(order.Uuid,string(json),0).Err()
	if err !=nil{
		panic(err.Error())
	}



}
func getProductById(id string) Product{
	response,err:= http.Get(productsUrl + "/products/" + id)
	if err!=nil{
		fmt.Printf("The HTTP reqiest failed with error %s\n",err)
	}
	data,_:=ioutil.ReadAll(response.Body)
	var product Product
	json.Unmarshal(data,&product)
	return product
}

func main(){
	in := make(chan []byte)

	StartConsuming(Connect(), in)
	for payload := range in {
		createOrder(payload)
		fmt.Println(string(payload))
	}
	
}
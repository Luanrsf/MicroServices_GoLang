package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"github.com/streadway/amqp"
	
)



type Product struct {
	Uuid    string  `json:"uuid"`
	Product string  `json:"product"`
	Price   float32 `json:"price,string"`
}

type Order struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	ProductId string `json:"product_id"`
}

var productsUrl string

func init() {
	productsUrl = os.Getenv("PRODUCT_URL")
}

func displayCheckout(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)
	response, err := http.Get(productsUrl + "/products/" + vars["id"])
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}

	data, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(data))

	var product Product
	json.Unmarshal(data, &product)

	t := template.Must(template.ParseFiles("templates/checkout.html"))
	t.Execute(w, product)
}
func Connect() *amqp.Channel {
	dsn := "amqp://" + os.Getenv("RABBITMQ_DEFAULT_USER") + ":" + os.Getenv("RABBITMQ_DEFAULT_PASS") + "@" + os.Getenv("RABBITMQ_DEFAULT_HOST") + ":" + os.Getenv("RABBITMQ_DEFAULT_PORT") + os.Getenv("RABBITMQ_DEFAULT_VHOST")

	conn, err := amqp.Dial(dsn)
	fmt.Println(dsn)
	fmt.Println(err)
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

func finish(w http.ResponseWriter, r *http.Request){
	var order Order
	order.Name = r.FormValue("name")
	order.Email = r.FormValue("email")
	order.Phone = r.FormValue("phone")
	order.ProductId = r.FormValue("product_id")
	data,_ := json.Marshal(order)
	fmt.Println([]byte(data))
	connection := Connect()
	Notify(data,"checkout_ex","",connection)
	w.Write([]byte("Processou!"))

}


func main() {
	r := mux.NewRouter()
	r.HandleFunc("/finish", finish)
	r.HandleFunc("/{id}", displayCheckout)
	http.ListenAndServe(":3336", r)
}
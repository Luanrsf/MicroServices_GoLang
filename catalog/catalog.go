package main
import (
	"fmt"
	"net/http"
	"os"
	"io/ioutil"
	"encoding/json"
	"html/template"
	"github.com/gorilla/mux"
)

type Product struct{
	Uuid string `json:"uuid"`
	Product string `json:"product"`
	Price float64 `json:"price,string"`
}
type Products struct{
	Products []Product
}
var productsUrl string

func init(){
	productsUrl = os.Getenv("PRODUCT_URL")
}

func loadProducts() []Product {
	response, err := http.Get(productsUrl + "/products")
	if err !=nil{
		fmt.Println("Erro de http")
	}
	data,_:= ioutil.ReadAll(response.Body)
	var products Products
	json.Unmarshal(data, &products)
	fmt.Println(string(data))
	return products.Products
}
func ListProducts(w http.ResponseWriter, r *http.Request){
	products:=loadProducts()
	t:=template.Must(template.ParseFiles("templates/catalog.html"))
	t.Execute(w,products)
}
func ShowProducts(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	response,err:= http.Get(productsUrl + "/products/" + vars["id"])
	if err!=nil{
		fmt.Printf("The HTTP reqiest failed with error %s\n",err)
	}
	data,_:=ioutil.ReadAll(response.Body)
	var product Product
	json.Unmarshal(data,&product)
	t:=template.Must(template.ParseFiles("templates/view.html"))
	t.Execute(w,product)

}

func main(){
	r:=mux.NewRouter()
	r.HandleFunc("/",ListProducts)
	r.HandleFunc("/products/{id}",ShowProducts)
	http.ListenAndServe(":3335",r)
}

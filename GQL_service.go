package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"bytes"
	"os"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

var data []*Product

func returnAllProducts() ([]*Product, error) {
	conn, err := http.Get("http://127.0.0.1:8081/get_products")
	if err != nil {
		fmt.Println("Error in connecting to product_service, error: ", err)
		return nil, nil
	}
	connBody, err := ioutil.ReadAll(conn.Body)
	if err != nil {
		fmt.Println("Error in reading data from product_service, error: ", err)
		return nil, nil
	}
	var dataFromSetv AllProucts
	err = json.Unmarshal(connBody, &dataFromSetv)
	if err != nil {
		fmt.Println("Error in unmarshal readed data from product_service, error: ", err)
		return nil, nil
	}
	data = dataFromSetv.ListProducts
	return data, nil
}

func returnAllOrdersByUser(userName string)([]*RedisListOrderProducts,error){
	mess := Mess{"OK",userName}
	jsonMess, err := json.Marshal(mess)
	if err != nil{
		fmt.Println("Error in marshal mess to json, error:",err)
		return nil, nil
	}
	dt := bytes.NewReader(jsonMess)
	conn, err := http.Post("http://127.0.0.1:8082/get_orders", "application/json", dt)
	if err != nil{
		fmt.Println("Error in getting order_server answer, error:",err)
		return nil, nil
	}
	connBody, err := ioutil.ReadAll(conn.Body)
	if err != nil{
		fmt.Println("Error in reading body order_server answer, error:",err)
		return nil, nil
	}
	var orderList []*RedisListOrderProducts
	err = json.Unmarshal(connBody,&orderList)
	if err != nil{
		fmt.Println("Error in unmarshal body order_server answer, error:",err)
		return nil, nil
	}
	return orderList, nil
}

type Product struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Descr string `json:"descr"`
}

type AllProucts struct {
	ListProducts []*Product `json:"list"`
}

type RedisListOrderProducts struct {
	Creator string              `json:"creator"`
	State   string              `json:"state"`
	List    []*RedisOrderProduct `json:"list"`
}

type RedisOrderProduct struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
	State  string `json:"state"`
}

type Mess struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}

func main() {
	// server := http.Server{Addr: "127.0.0.1:8084"}
	fmt.Println("Start GQL_service")
	defer fmt.Println("Stop GQL_service")
	productType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Product",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(*Product); ok {
						return product.Id, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(*Product); ok {
						return product.Name, nil
					}
					return nil, nil
				},
			},
			"desciption": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(*Product); ok {
						return product.Descr, nil
					}
					return nil, nil
				},
			},
		},
	})
	orderProductType := graphql.NewObject(graphql.ObjectConfig{
		Name: "list",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(*RedisOrderProduct); ok {
						return product.Name, nil
					}
					return nil, nil
				},
			},
			"number": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(*RedisOrderProduct); ok {
						return product.Number, nil
					}
					return nil, nil
				},
			},
			"state": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(*RedisOrderProduct); ok {
						return product.State, nil
					}
					return nil, nil
				},
			},
		},
	})
	orderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"creator": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if orderProduct, ok := p.Source.(*RedisListOrderProducts); ok {
						return orderProduct.Creator, nil
					}
					return nil, nil
				},
			},
			"state": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if orderProduct, ok := p.Source.(*RedisListOrderProducts); ok {
						return orderProduct.State, nil
					}
					return nil, nil
				},
			},
			"list": &graphql.Field{
				Type: graphql.NewList(orderProductType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if orderProduct, ok := p.Source.(*RedisListOrderProducts); ok {
						return orderProduct.List, nil
					}
					return nil, nil
				},
			},
		},
	})
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"products": &graphql.Field{
				Type: graphql.NewList(productType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return returnAllProducts()
				},
			},
			"order": &graphql.Field{
				Type: graphql.NewList(orderType),
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return returnAllOrdersByUser(p.Args["name"].(string))
				},
			},
		},
	})

	fmt.Println("Creating shcema for product_service")
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Schema was successful created")

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: false,
	})

	http.Handle("/getAllProducts", h)
	fmt.Println("GQL_service run")
	go func() {
		err = http.ListenAndServe(":8084", nil)
		fmt.Println("ERROR:", err)
	}()
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if text == "exit\n" {
			return
		}
	}
}

package main

import (
	"fmt"
)
import "os"
import "bufio"
import "context"
import "net/http"
import "io/ioutil"
import "encoding/json"
import "github.com/google/uuid"
import "github.com/go-redis/redis/v8"
import "time"
import "bytes"
import "strconv"
import ampq "github.com/rabbitmq/amqp091-go"

// import "reflect"

var ctx = context.Background()
var client *redis.Client
var ch *ampq.Channel
var q ampq.Queue

type OrderProduct struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
}

type RedisOrderProduct struct {
	Name   string `redis:"name" json:"name"`
	Number int    `redis:"number" json:"number"`
	State  string `redis:"state" json:"state"`
}

type RedisListOrderProducts struct {
	Creator string              `redis:"creator" json:"creator"`
	State   string              `redis:"state" json:"state"`
	List    []RedisOrderProduct `redis:"list" json:"list"`
}

type ListOrderProducts struct {
	Creator string         `json:"creator"`
	List    []OrderProduct `json:"list"`
}

type Mess struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}

type Product struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

func packageAndSendMess(w http.ResponseWriter, ms Mess) {
	outputData, err := json.Marshal(ms)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Ошибка в конвертации в json", 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(outputData)
}

func handlerAddOrder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Begin creating order")
	var ms Mess
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка чтения тела соощения"
		packageAndSendMess(w, ms)
		return
	}
	var info ListOrderProducts
	err = json.Unmarshal(body, &info)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка конвертирования из json тела сообщения"
		packageAndSendMess(w, ms)
		return
	}
	key := "base"
	val, err := client.Get(ctx, key).Result()
	if val == "" {
		fmt.Println("KeyDB server not available")
		ms.Status = "ERROR"
		ms.Info = "База данных недоступна"
		packageAndSendMess(w, ms)
		return
	}
	var template string
	template = info.Creator
	template += ":orders/"
	id := uuid.New()
	template += uuid.Must(id, err).String()
	var redisInfo RedisListOrderProducts
	redisInfo.Creator = info.Creator
	redisInfo.State = "in progress"
	counter := 0
	for _, prod := range info.List {
		redisProd := RedisOrderProduct{prod.Name, prod.Number, "in progress"}
		productForInventory := Product{prod.Name, prod.Number}
		prodInve, err := json.Marshal(productForInventory)
		if err != nil {
			fmt.Println(err)
			ms.Status = "ERROR"
			ms.Info = "Ошибка в конвертации в json элемента заказа для микросервиса инвентаря"
			packageAndSendMess(w, ms)
			return
		}
		rs := bytes.NewReader(prodInve)
		conn, err := http.Post("http://127.0.0.1:8083/sub_inv", "application/json", rs)
		if err != nil {
			fmt.Println(err)
			ms.Status = "ERROR"
			ms.Info = "Ошибка в кв подключении к микросервису инвентаря"
			packageAndSendMess(w, ms)
			return
		}
		connBody, err := ioutil.ReadAll(conn.Body)
		if err != nil {
			fmt.Println(err)
			ms.Status = "ERROR"
			ms.Info = "Ошибка в чтении тела ответа микросервиса инвентаря"
			packageAndSendMess(w, ms)
			return
		}
		var connMs Mess
		err = json.Unmarshal(connBody, &connMs)
		if err != nil {
			fmt.Println(err)
			ms.Status = "ERROR"
			ms.Info = "Ошибка в конвертации из json  ответа с микросервиса инвентаря"
			packageAndSendMess(w, ms)
			return
		}
		if connMs.Status == "OK" && connMs.Info == "done" {
			redisProd.State = connMs.Info
			counter++
		} else if connMs.Status == "ERROR" {
			fmt.Println(err)
			ms.Status = "ERROR"
			ms.Info = "В заказе находится несуществующий товар!"
			packageAndSendMess(w, ms)
			return
		}
		redisInfo.List = append(redisInfo.List, redisProd)
	}
	if counter == len(redisInfo.List) {
		redisInfo.State = "done"
	}
	ds, err := json.Marshal(redisInfo)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в конвертации в json сформированно заказа"
		packageAndSendMess(w, ms)
		return
	}
	err = client.Set(ctx, template, []byte(ds), time.Minute*5).Err()
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в отправке сформированного заказа в БД"
		packageAndSendMess(w, ms)
		return
	}
	for _, product := range redisInfo.List {
		if product.State == "done" {
			Info := "Зарезервировано <" + strconv.Itoa(product.Number) + "><" + product.Name + "> для заказа <" + uuid.Must(id, err).String() + ">"
			newctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = ch.PublishWithContext(newctx,
				"",
				q.Name,
				false,
				false,
				ampq.Publishing{
					ContentType: "text/plain",
					Body:        []byte(Info),
				},
			)
			defer cancel()
		}
	}
	fmt.Println("Creating order was successful")
	ms.Status = "OK"
	ms.Info = "заказ сформирован"
	packageAndSendMess(w, ms)
}

func handlerGetOrders(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Begin getting orders")
	var ms, input Mess
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в чтении тела запроса"
		packageAndSendMess(w, ms)
		return
	}
	err = json.Unmarshal(requestBody, &input)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в конвертации из json тела запроса"
		packageAndSendMess(w, ms)
		return
	}
	key := "base"
	val, err := client.Get(ctx, key).Result()
	if val == "" {
		fmt.Println("KeyDB server not available")
		ms.Status = "ERROR"
		ms.Info = "База данных недоступна"
		packageAndSendMess(w, ms)
		return
	}
	results, err := client.Do(ctx, "keys", input.Info+"*").Result()
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Не был найден пользователь <" + input.Info + ">"
		packageAndSendMess(w, ms)
		return
	}
	var orderList []RedisListOrderProducts
	if rec, ok := results.([]interface{}); ok {
		for _, recc := range rec {
			if reccc, ok := recc.(string); ok {
				val, err = client.Get(ctx, reccc).Result()
				if err != nil {
					fmt.Println(err)
					ms.Status = "ERROR"
					ms.Info = "Ошибка в чтении заказа <" + reccc + ">"
					packageAndSendMess(w, ms)
					return
				}
				var newOrder RedisListOrderProducts
				err = json.Unmarshal([]byte(val), &newOrder)
				if err != nil {
					fmt.Println(err)
					ms.Status = "ERROR"
					ms.Info = "Ошибка в распаковке содержимого заказа <" + reccc + ">"
					packageAndSendMess(w, ms)
					return
				}
				orderList = append(orderList, newOrder)
			}
		}
	}
	res, err := json.Marshal(orderList)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в конвертации ответа в json"
		packageAndSendMess(w, ms)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
	fmt.Println("Getting orders correctly complited")
}

func main() {
	fmt.Println("Orders server start to work")
	defer fmt.Println("Orders server stop working")
	fmt.Println("Connecting to keyDB server")
	defer fmt.Println("Close connection to keyDB server")
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	key := "base"
	val, err := client.Get(ctx, key).Result()
	if val == "" {
		fmt.Println("KeyDB server not available")
		fmt.Println(err)
		return
	}
	fmt.Println("Connecting to DB was successful")
	fmt.Println("Connecting to RabbitMQ")
	defer fmt.Println("Close connection to RabbitMQ")
	conn, err := ampq.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("RabbitMQ is not available")
		fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("Connecting to RabbitMQ was successful")
	fmt.Println("Creating RMQ channel")
	ch, err = conn.Channel()
	if err != nil {
		fmt.Println("Error in creating RMQ channel")
		fmt.Println(err)
		return
	}
	fmt.Println("Creating RMQ channel was successful")
	defer ch.Close()
	fmt.Println("Creating RMQ queue")
	q, err = ch.QueueDeclare(
		"Inventory", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error in creating RMQ queue")
		return
	}
	fmt.Println("Creating RMQ queue was successful")
	http.HandleFunc("/add_order", handlerAddOrder)
	http.HandleFunc("/get_orders", handlerGetOrders)
	go http.ListenAndServe(":8082", nil)
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if text == "exit\n" {
			return
		}
	}

}

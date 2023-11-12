package main
import "fmt"
import "os"
import "bufio"
import "net/http"
import "database/sql"
import _ "github.com/lib/pq"
import "encoding/json"

var DB *sql.DB

type Product struct{
	Id int `json:"id"`
	Name string `json:"name"`
	Descr string `json:"descr"`
}

type AllProucts struct{
	ListProducts []Product `json:"list"`
}

func handler(w http.ResponseWriter, _ *http.Request){
	DBquestion := "SELECT id,name,descr FROM products"
	fmt.Println("Sending '", DBquestion, "' to DB")
	info, err := DB.Query(DBquestion)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Info was gotten")
	var counter int = 0
	message := AllProucts{}
	for info.Next(){
		prod := Product{}
		err := info.Scan(&prod.Id,&prod.Name,&prod.Descr)
		if err != nil {
			fmt.Println(err)
			continue
		}
		message.ListProducts = append(message.ListProducts,prod)
		counter++
	}
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(data)
	fmt.Println("Was send ", counter, " elems")
}

func CloseDB(){
	DB.Close()
	fmt.Println("Connection with DB was closed")
}

func main(){
	fmt.Println("Products server begin to work")
	defer fmt.Println("Products server stop working")
	dbConnStr := "user=postgres password=admin dbname=productsDB sslmode=disable"
	db, err := sql.Open("postgres",dbConnStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	DB = db
	fmt.Println("Products server connect to DB")
	defer CloseDB()
	http.HandleFunc("/get_products", handler)
	go http.ListenAndServe("127.0.0.1:8081", nil)
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if text == "exit\n"{
			return
		}
	}
}
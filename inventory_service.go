package main
import "fmt"
import "os"
import "bufio"
import "net/http"
import "database/sql"
import _ "github.com/lib/pq"
import "encoding/json"
import "io/ioutil"
// import "strconv"

var DB *sql.DB

type Product struct{
	Name string `json:"name"`
	Quantity int `json:"quantity"`
}

type ProdList struct {
	List []Product `json:"list"`
}

type Mess struct{
	Status string `json:"status"`
	Info string `json:"info"`
}

func packageAndSendMess(w http.ResponseWriter, ms Mess){
	outputData, err := json.Marshal(ms)
	if err != nil {
		fmt.Println(err)
		http.Error(w,"Ошибка в конвертации в json",500)
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(outputData)
}

func handlerSubInv(w http.ResponseWriter, r *http.Request){
	fmt.Println("Begin sub inventory")
	var ms Mess
	responceBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка чтения тела запроса"
		packageAndSendMess(w,ms)
		return
	}
	var inputProduct Product
	err = json.Unmarshal(responceBody,&inputProduct)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в конвертации json в теле сообщения"
		packageAndSendMess(w,ms)
		return
	}
	fmt.Println("Getting list of existing products")
	info, err := DB.Query("SELECT name,quantity FROM products")
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в подключении к БД"
		packageAndSendMess(w,ms)
		return
	}
	fmt.Println("Info was gotten from DB")
	var existingPoroductdList ProdList
	counter := 0
	for info.Next(){
		prod := Product{}
		err := info.Scan(&prod.Name,&prod.Quantity)
		if err != nil {
			fmt.Println(err)
			continue
		}
		existingPoroductdList.List = append(existingPoroductdList.List,prod)
		counter++
	}
	fmt.Println("Was gotten <",counter,"> rows")
	fmt.Println("Check to existing product")
	flag_exist := false
	var existProduct Product
	for _, existProd := range existingPoroductdList.List {
		if inputProduct.Name == existProd.Name {
			flag_exist = true
			existProduct = existProd
			break
		}
	}
	if flag_exist == false {
		fmt.Println("Input unexisting product <",inputProduct.Name,">!")
		ms.Status = "ERROR"
		ms.Info = "Был получен несуществующий товар <"+inputProduct.Name+">!"
		packageAndSendMess(w,ms)
		return
	}
	fmt.Println("Input product is existing")
	if inputProduct.Quantity < 0 {
		ms.Status = "ERROR"
		ms.Info = "Число товаров меньше 0!"
		packageAndSendMess(w,ms)
		return
	} else if inputProduct.Quantity > existProduct.Quantity {
		ms.Info = "in progress"
		ms.Status = "OK"
	} else {
		existProduct.Quantity -= inputProduct.Quantity
		ms.Info = "done"
		ms.Status = "OK"
	}
	fmt.Println("Updating data in DB")
	_, err = DB.Exec("UPDATE products SET quantity = $1 WHERE name = $2",existProduct.Quantity,existProduct.Name)
	if err != nil {
		fmt.Println(err)
		ms.Status = "ERROR"
		ms.Info = "Ошибка в обновлении данных в БД"
		packageAndSendMess(w,ms)
		return
	}
	fmt.Println("Updating data in DB correctly complited")
	packageAndSendMess(w,ms)
	fmt.Println("Sub inventory was correctly comlited")
}

func CloseDB(){
	DB.Close()
	fmt.Println("Connection with DB was closed")
}

func main(){
	fmt.Println("Inventory server begin to work")
	defer fmt.Println("Inventory server stop working")
	dbConnStr := "user=postgres password=admin dbname=productsDB sslmode=disable"
	db, err := sql.Open("postgres",dbConnStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	DB = db
	fmt.Println("Inventory server connect to DB")
	defer CloseDB()
	http.HandleFunc("/sub_inv", handlerSubInv)
	go http.ListenAndServe("127.0.0.1:8083", nil)
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if text == "exit\n"{
			return
		}
	}	
}
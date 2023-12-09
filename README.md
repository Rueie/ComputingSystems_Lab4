<p align="center">
  <img src="https://github.com/Rueie/ComputingSystems_Lab4/assets/99559651/0a96d28f-ec7d-43fc-a173-12b1b2df23c0"/>
</p>

# product_service.go
## Назначение
Позволяет получать список всех товаров с их описанием

## С чем связана
* БД PSQL (productsDB)
* Микросервис GQL_service.go (RestAPI)

## Поддерживаемые запросы
* **GET** http://127.0.0.1:8082/get_orders

**Результат запроса:**
```json
{
    "list": [
        {
            "id": 9,
            "name": "Табуретка",
            "descr": "Прекрасная табуретка"
        },
        {
            "id": 10,
            "name": "Эксклюзивная подушка",
            "descr": "Прекрасная эксклюзивная подушка"
        },
        {
            "id": 8,
            "name": "Шкаф",
            "descr": "Прекрасный шкаф"
        },
        {
            "id": 7,
            "name": "Стол",
            "descr": "Прекрасный стол"
        },
        {
            "id": 6,
            "name": "Стул",
            "descr": "Прекрасный стул"
        }
    ]
}
```
# notification_service.go
## Назначение
Уведомляет работников скалада о резервации товара в следующей форме: <br>
'Зарезервированно <**Число товара**><**Название товара**> для заказа <**UUID заказа**>' <br>
Например: <br>
```
2023/11/12 17:48:50 Received a message: Зарезервировано <5><Стол> для заказа <d3a99ecc-7748-4703-a56d-382bc2bc9243>
2023/11/12 17:48:50 Received a message: Зарезервировано <4><Стул> для заказа <d3a99ecc-7748-4703-a56d-382bc2bc9243>
```

## С чем связана
* order_service.go (RMQ)

## Поддерживаемые запросы
Отсутствуют

# inventory_service.go
## Назначение
Позволяет резервировать товары для заказов путём проверки их наличия на складе, вычитания числа зарезервированных товаров и возврат результат резервирования

## С чем связана
* БД PSQL (productsDB)
* order_service.go (RestAPI)

## Поддерживаемые запросы
* **POST** http://127.0.0.1:8083/sub_inv <br>
Запрос тела в формате **raw**: <br>
```json
{
    "name" : "название товара",
    "quantity" :  1
}
```
### Реакция на некорректные данные
* Несуществующий товар <br>
Запрос: **GET** http://127.0.0.1:8083/sub_inv <br>
Запрос тела в формате **raw**: <br>
```json
{
    "name" : "стол",
    "quantity" : 4
}
```
**Результат запроса**:
```json
{
    "status": "ERROR",
    "info": "Был получен несуществующий товар <стол>!"
}
```
* Невозможное число товаров <br>
Запрос: **POST** http://127.0.0.1:8083/sub_inv <br>
Запрос тела в формате **raw**: <br>
```json
{
    "name" : "Стол",
    "quantity" : -1
}
```
**Результат запроса**:
```json
{
    "status": "ERROR",
    "info": "Число товаров меньше 0!"
}
```
### Реакция на корректные запросы
* Число товаров больше, чем есть на складе <br>
Запрос: **POST** http://127.0.0.1:8083/sub_inv <br>
Запрос тела в формате **raw**: <br>
```json
{
    "name" : "Стол",
    "quantity" : 1000
}
```
**Таблица products до выполнения запроса**
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Табуретка|Прекрасная табуретка|9|1000|
|Эксклюзивная подушка|Прекрасная эксклюзивная подушка|10|1000|
|Шкаф|Прекрасный шкаф|8|1000|
|Стул|Прекрасный стул|6|912|
|Стол|Прекрасный стол|7|882|

**Результат запроса**:
```json
{
    "status": "OK",
    "info": "in progress"
}
```
**Таблица products после выполнения запроса**
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Табуретка|Прекрасная табуретка|9|1000|
|Эксклюзивная подушка|Прекрасная эксклюзивная подушка|10|1000|
|Шкаф|Прекрасный шкаф|8|1000|
|Стул|Прекрасный стул|6|912|
|Стол|Прекрасный стол|7|882|

* Необходимое число товаров есть на складе <br>
Запрос: **POST** http://127.0.0.1:8083/sub_inv <br>
Запрос тела в формате **raw**: <br>
```json
{
    "name" : "Стол",
    "quantity" : 4
}
```
**Таблица products до выполнения запроса**
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Табуретка|Прекрасная табуретка|9|1000|
|Эксклюзивная подушка|Прекрасная эксклюзивная подушка|10|1000|
|Шкаф|Прекрасный шкаф|8|1000|
|Стул|Прекрасный стул|6|912|
|Стол|Прекрасный стол|7|882|

**Результат запроса**:
```json
{
    "status": "OK",
    "info": "done"
}
```
**Таблица products после выполнения запроса**
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Табуретка|Прекрасная табуретка|9|1000|
|Эксклюзивная подушка|Прекрасная эксклюзивная подушка|10|1000|
|Шкаф|Прекрасный шкаф|8|1000|
|Стул|Прекрасный стул|6|912|
|Стол|Прекрасный стол|7|878|


# order_service.go
## Назначение 
Формирования заказов и просмотр текущего списка заказов подьзователя по его имени

## С чем связана
* notification_service.go (RMQ)
* inventory_service.go (RestAPI)
* KeyDB
* GQL_service.go (RestAPI)

## Поддерживаемые запросы
* **POST** http://127.0.0.1:8082/add_order <br>
Запрос тела в формате **raw**: <br>
```json
{
    "creator": "создатель заказа",
    "list" : [
        {
            "name" : "название товара 1",
            "number" : 5100
        },
        {
            "name" : "название товара 2",
            "number" : 5
        },
        {
            "name" : "название товара 3",
            "number" : 4
        }
    ]
}
```
* **POST** http://127.0.0.1:8082/get_orders <br>
Запрос тела в формате **raw**: <br>
```json
{
    "status":"OK",
    "info":"Rueie"
}
```
### Реакция на некорректные запросы
* Несущетсвующий товар или невозможное число товаров (т.е. получили ошибку от inventory_service.go) <br>
Запрос: **POST** http://127.0.0.1:8082/add_order <br>
Запрос тела в формате **raw**: <br>
```json
{
    "creator": "Rueie",
    "list" : [
        {
            "name" : "что-то",
            "number" : 5100
        },
        {
            "name" : "Стол",
            "number" : 5
        },
        {
            "name" : "Стул",
            "number" : 4
        }
    ]
}
```
**Результат запроса:**
```json
{
    "status": "ERROR",
    "info": "В заказе находится несуществующий товар!"
}
```
Таким же будет и эффет на следующее тело запроса: <br>
```json
{
    "creator": "Rueie",
    "list" : [
        {
            "name" : "Шкаф",
            "number" : -1
        },
        {
            "name" : "Стол",
            "number" : 5
        },
        {
            "name" : "Стул",
            "number" : 4
        }
    ]
}
```
* Нет имени заказчика <br>
Запрос: **POST** http://127.0.0.1:8082/get_orders <br>
Запрос тела в формате **raw**: <br>
```json
{
    "status":"ERROR",
    "info":""
}
```
**Результат запроса:**
```json
{
    "status": "ERROR",
    "info": "Ошибка в распаковке содержимого заказа <base>"
}
```
* Нет поля имени заказчика <br>
Запрос: **POST** http://127.0.0.1:8082/get_orders <br>
Запрос тела в формате **raw**: <br>
```json
{
    "status":"ERROR",
}
```
**Результат запроса:**
```json
{
    "status": "ERROR",
    "info": "Ошибка в конвертации из json тела запроса"
}
```
### Реакция на корректные запросы
* Корректный заказ <br>
Запрос: **POST** http://127.0.0.1:8082/add_order <br>
Запрос тела в формате **raw**: <br>
```json
{
    "creator": "Rueie",
    "list" : [
        {
            "name" : "Шкаф",
            "number" : 5100
        },
        {
            "name" : "Стол",
            "number" : 5
        },
        {
            "name" : "Стул",
            "number" : 4
        }
    ]
}
```
**Таблица products до выполнения запроса**
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Табуретка|Прекрасная табуретка|9|1000|
|Эксклюзивная подушка|Прекрасная эксклюзивная подушка|10|1000|
|Шкаф|Прекрасный шкаф|8|1000|
|Стол|Прекрасный стол|7|863|
|Стул|Прекрасный стул|6|900|

**Данные по заказам пользователя в KeyDB:** <br>
```
(empty array)
```
**Результат запроса:**
```json
{
    "status": "OK",
    "info": "заказ сформирован"
}
```
**Таблица products после выполнения запроса**
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Табуретка|Прекрасная табуретка|9|1000|
|Эксклюзивная подушка|Прекрасная эксклюзивная подушка|10|1000|
|Шкаф|Прекрасный шкаф|8|1000|
|Стол|Прекрасный стол|7|858|
|Стул|Прекрасный стул|6|896|

**Данные по заказам пользователя в KeyDB:** <br>
```
1) "Rueie:orders/04ee99a9-9dc5-47d9-98d5-11b934f18afe"
```
**Содержимое заказа *04ee99a9-9dc5-47d9-98d5-11b934f18afe* пользователя *Rueie*:**
```
"{\"creator\":\"Rueie\",\"state\":\"in progress\",\"list\":[{\"name\":\"\xd0\xa8\xd0\xba\xd0\xb0\xd1\x84\",\"number\":5100,\"state\":\"in progress\"},{\"name\":\"\xd0\xa1\xd1\x82\xd0\xbe\xd0\xbb\",\"number\":5,\"state\":\"done\"},{\"name\":\"\xd0\xa1\xd1\x82\xd1\x83\xd0\xbb\",\"number\":4,\"state\":\"done\"}]}"
```
GQL_service.go - сервис, который реалзиует возможность получения данных с помощью GraphQL на Golang
* Список заказов пользователя <br>
Запрос: **POST** http://127.0.0.1:8082/get_orders <br>
Запрос тела в формате **raw**: <br>
```json
{
    "status":"OK",
    "info":"Rueie"
}
```
**Результат запроса:** <br>
```json
[
    {
        "creator": "Rueie",
        "state": "done",
        "list": [
            {
                "name": "Шкаф",
                "number": 8,
                "state": "done"
            },
            {
                "name": "Стол",
                "number": 5,
                "state": "done"
            },
            {
                "name": "Стул",
                "number": 4,
                "state": "done"
            }
        ]
    },
    {
        "creator": "Rueie",
        "state": "in progress",
        "list": [
            {
                "name": "Шкаф",
                "number": 5100,
                "state": "in progress"
            },
            {
                "name": "Стол",
                "number": 5,
                "state": "done"
            },
            {
                "name": "Стул",
                "number": 4,
                "state": "done"
            }
        ]
    },
    {
        "creator": "Rueie",
        "state": "in progress",
        "list": [
            {
                "name": "Шкаф",
                "number": 5100,
                "state": "in progress"
            },
            {
                "name": "Стол",
                "number": 5,
                "state": "done"
            },
            {
                "name": "Стул",
                "number": 4,
                "state": "done"
            }
        ]
    }
]
```
* Выбран пользователь без заказов <br>
Запрос: **POST** http://127.0.0.1:8082/get_orders <br>
Запрос тела в формате **raw**: <br>
```json
{
    "status":"OK",
    "info":"Кто-то"
}
```
**Результат запроса:** <br>
```json
null
```

# GQL_service.go
## Назначение
Формировать json представление данных по схеме из других сервисов

## С чем связан
* product_service.go (RestAPI)
* order_service.go (RestAPI)

## Схема GraphQL
Товар из списка товаров: <br>
```GraphQL
Product {
  "id": Int
  "name": String
  "desciption": String
}
```
Товар из заказа: <br>
```GraphQL
list{
  "name": String
  "number": Int
  "state": String
}
```
Заказ: <br>
```graphQL
Order{
  "creator": String
  "state": String
  "list": list
}
```
Доступные операции в Query: <br>
```GraphQL
Query{
  "products": []*Product
  "order"("name": String): []*Order
}
```
Создание заказа:
```GraphQL
Mutation {
  createOrder (creator: "Имя заказчика", productNames: ["Название товара", "Название товара"], productNumber: [Число единиц товара, Число единиц товара])
}
```
## Поддерживаемые запросы
* **POST** http://127.0.0.1:8084/getAllProducts <br>
Запрос тела в формате **GraphQL**: <br>
```GraphQL
query {
    products {
        id
        name
        desciption
    }
    order(name:"Имя заказчика") {
        creator
        state
        list {
            name
            number
            state
        }
    }
}
mutation {
  createOrder (creator: "Имя заказчика", productNames: ["Название товара", "Название товара"], productNumber: [Число единиц товара, Число единиц товара])
}
```

## Примеры запросов
* Запрос товаров из product_service.go <br>
Запрос: **POST** http://127.0.0.1:8084/getAllProducts <br>
Запрос тела в формате **GraphQL**: <br>
```GraphQL
query {
    products {
        id
        name
        desciption
    }
}
```
**Результат запроса:**
```json
{
    "data": {
        "products": [
            {
                "desciption": "Прекрасная табуретка",
                "id": 9,
                "name": "Табуретка"
            },
            {
                "desciption": "Прекрасная эксклюзивная подушка",
                "id": 10,
                "name": "Эксклюзивная подушка"
            },
            {
                "desciption": "Прекрасный шкаф",
                "id": 8,
                "name": "Шкаф"
            },
            {
                "desciption": "Прекрасный стол",
                "id": 7,
                "name": "Стол"
            },
            {
                "desciption": "Прекрасный стул",
                "id": 6,
                "name": "Стул"
            }
        ]
    }
}
```
* Запрос заказов для пользователя **Rueie** <br>
Запрос: **POST** http://127.0.0.1:8084/getAllProducts <br>
Запрос тела в формате **GraphQL**: <br>
```GraphQL
query {
    order(name:"Rueie") {
        creator
        state
        list {
            name
            number
            state
        }
    }
}
```
**Результат запроса:**
```json
{
    "data": {
        "order": [
            {
                "creator": "Rueie",
                "list": [
                    {
                        "name": "Шкаф",
                        "number": 1000,
                        "state": "in progress"
                    },
                    {
                        "name": "Стол",
                        "number": 5,
                        "state": "done"
                    },
                    {
                        "name": "Стул",
                        "number": 4,
                        "state": "done"
                    }
                ],
                "state": "in progress"
            },
            {
                "creator": "Rueie",
                "list": [
                    {
                        "name": "Шкаф",
                        "number": 8,
                        "state": "done"
                    },
                    {
                        "name": "Стол",
                        "number": 5,
                        "state": "done"
                    },
                    {
                        "name": "Стул",
                        "number": 4,
                        "state": "done"
                    }
                ],
                "state": "done"
            }
        ]
    }
}
```
* Объединённый пример с некоторыми убранными полями <br>
Запрос: **POST** http://127.0.0.1:8084/getAllProducts <br>
Запрос тела в формате **GraphQL**: <br>
```GraphQL
query {
    products {
        # id
        name
        desciption
    }
    order(name:"Rueie") {
        creator
        state
        list {
            name
            # number
            state
        }
    }
}
```
**Результат запроса:**
```json
{
    "data": {
        "order": [
            {
                "creator": "Rueie",
                "list": [
                    {
                        "name": "Шкаф",
                        "state": "in progress"
                    },
                    {
                        "name": "Стол",
                        "state": "done"
                    },
                    {
                        "name": "Стул",
                        "state": "done"
                    }
                ],
                "state": "in progress"
            },
            {
                "creator": "Rueie",
                "list": [
                    {
                        "name": "Шкаф",
                        "state": "done"
                    },
                    {
                        "name": "Стол",
                        "state": "done"
                    },
                    {
                        "name": "Стул",
                        "state": "done"
                    }
                ],
                "state": "done"
            }
        ],
        "products": [
            {
                "desciption": "Прекрасная табуретка",
                "name": "Табуретка"
            },
            {
                "desciption": "Прекрасная эксклюзивная подушка",
                "name": "Эксклюзивная подушка"
            },
            {
                "desciption": "Прекрасный шкаф",
                "name": "Шкаф"
            },
            {
                "desciption": "Прекрасный стол",
                "name": "Стол"
            },
            {
                "desciption": "Прекрасный стул",
                "name": "Стул"
            }
        ]
    }
}
```
* Запрос заказов для несуществующего пользователя или для пользователя, у которого нет заказов <br>
Запрос: **POST** http://127.0.0.1:8084/getAllProducts <br>
Запрос тела в формате **GraphQL**: <br>
```GraphQL
query {
    products {
        id
        name
        desciption
    }
    order(name:"Rueie") {
        creator
        state
        list {
            name
            number
            state
        }
    }
}
```
**Результат запроса:**
```json
{
    "data": {
        "order": [],
        "products": [
            {
                "desciption": "Прекрасная табуретка",
                "id": 9,
                "name": "Табуретка"
            },
            {
                "desciption": "Прекрасная эксклюзивная подушка",
                "id": 10,
                "name": "Эксклюзивная подушка"
            },
            {
                "desciption": "Прекрасный шкаф",
                "id": 8,
                "name": "Шкаф"
            },
            {
                "desciption": "Прекрасный стол",
                "id": 7,
                "name": "Стол"
            },
            {
                "desciption": "Прекрасный стул",
                "id": 6,
                "name": "Стул"
            }
        ]
    }
}
```
* Запрос на создание заказа <br>
Запрос: **POST** http://127.0.0.1:8084/getAllProducts <br>
Запрос тела в формате **GraphQL**: <br>
```GraphQL
mutation {
  createOrder (creator: "rueie", productNames: ["Как выучить с++ за месяц", "100 рецептов для новичков"], productNumber: [1, 120])
}
```
__БД товаров до получения запроса__
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Как выучить с++ за месяц|Книга про базовые знания в с++|1|100|
|100 рецептов для новичков|100 различных рецептов для начианющих готовить|2|100|
|Война и мир|Классика жанра|3|100|
|Сборник анекдотов|Сборник анекдотов квн 2000-х годов|4|100

__БД заказов до получения запроса__
```
172.20.0.2:30002> keys rueie*
(empty array)
```
__БД товаров после получения запроса__
| name | descr | id | quantity|
| :---: | :---: | :---: | :---: |
|Как выучить с++ за месяц|Книга про базовые знания в с++|1|99|
|100 рецептов для новичков|100 различных рецептов для начианющих готовить|2|100|
|Война и мир|Классика жанра|3|100|
|Сборник анекдотов|Сборник анекдотов квн 2000-х годов|4|100

__БД заказов после получения запроса__
```
172.20.0.2:30002> keys rueie*
1) "rueie:orders/12877bff-1813-4fc5-9ec5-ab3b3516c9be"
172.20.0.2:30002> get rueie:orders/12877bff-1813-4fc5-9ec5-ab3b3516c9be
"{\"creator\":\"rueie\",\"state\":\"in progress\",\"list\":[{\"name\":\"\xd0\x9a\xd0\xb0\xd0\xba \xd0\xb2\xd1\x8b\xd1\x83\xd1\x87\xd0\xb8\xd1\x82\xd1\x8c \xd1\x81++ \xd0\xb7\xd0\xb0 \xd0\xbc\xd0\xb5\xd1\x81\xd1\x8f\xd1\x86\",\"number\":1,\"state\":\"done\"},{\"name\":\"100 \xd1\x80\xd0\xb5\xd1\x86\xd0\xb5\xd0\xbf\xd1\x82\xd0\xbe\xd0\xb2 \xd0\xb4\xd0\xbb\xd1\x8f \xd0\xbd\xd0\xbe\xd0\xb2\xd0\xb8\xd1\x87\xd0\xba\xd0\xbe\xd0\xb2\",\"number\":120,\"state\":\"in progress\"}]}"
```
__Запрос на получение данных о заказах заказчика__
```GraphQL
query {
    order(name:"rueie") {
        creator
        state
        list {
            name
            number
            state
        }
    }
}
```
__Результат запроса__
```json
{
    "data": {
        "order": [
            {
                "creator": "rueie",
                "list": [
                    {
                        "name": "Как выучить с++ за месяц",
                        "number": 1,
                        "state": "done"
                    },
                    {
                        "name": "100 рецептов для новичков",
                        "number": 120,
                        "state": "in progress"
                    }
                ],
                "state": "in progress"
            }
        ]
    }
}
```

# PSQL
## Назначение
Хранит товары и их количество
## База данных productsDB
### Таблицы
* products

#### Таблица products
| Название поля | Тип поля |
| :----: | :----: |
| name | text |
| descr | text |
| id | integer |
| quantity | numeric(5.0) |

# RMQ
## Очереди
Название очерди: **Inventory**

# KeyDB
## Назначение
Хранит состав заказов 5 минут, после чего заказ уничтожается. <br>
### Стркутура заказа
```json
{
  "creator": "кто создал заказ",
  "status": "статус заказа (done/in progress)"
  "list": [
    {
      "name": "название товара 1",
      "status": "статус заказа (done/in progress)",
      "quantity": 5
    },
    {
      "name": "название товара n",
      "status": "статус заказа (done/in progress)",
      "quantity": 5
    },
  ]
}
```
### Как хранятся заказы
Все заказы хранятся в следующем виде: <br>
**Имя_создателя**:orders/**UUID_заказа** <br>
Так, например, после внесения команды: 
```
keys Rueie*
```
мы получаем следующий список заказов пользователя **Rueie**: <br>
```
1) "Rueie:orders/9b428e95-7650-4bc3-8ba7-b82f23927637"
2) "Rueie:orders/17728171-1260-45f0-91d4-ba5c877e784c"
```

Где последний заказ можно получить с помощью команды: <br>
```
get Rueie:orders/17728171-1260-45f0-91d4-ba5c877e784c
```
имеет следующий вид: <br>
```
"{\"creator\":\"Rueie\",\"state\":\"in progress\",\"list\":[{\"name\":\"\xd0\xa8\xd0\xba\xd0\xb0\xd1\x84\",\"number\":5100,\"state\":\"in progress\"},{\"name\":\"\xd0\xa1\xd1\x82\xd0\xbe\xd0\xbb\",\"number\":5,\"state\":\"done\"},{\"name\":\"\xd0\xa1\xd1\x82\xd1\x83\xd0\xbb\",\"number\":4,\"state\":\"done\"}]}"
```

# RestAPI
Реализовывалось с помощью библиотеки ***net/http*** языка goland. <br>
Все сообщения передаются в виде формата **json**, которые можно увидеть в ранее описанных сервисах.

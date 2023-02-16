package main

import (
	data "awesomeAPI/src/dataBase"
	_ "github.com/alexbrainman/odbc"
	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	router.GET("/getItems", data.GetItems)
	router.GET("/getItemsCount", data.GetItemCount)

	router.Run("localhost:8080")

	//fmt.Println(items)
}

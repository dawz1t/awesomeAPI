package dataBase

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/charmap"
	"log"
	"net/http"
	"strconv"
)

type Item struct {
	Id                 uint32
	Brand              string
	ProductName        string
	ProductGroup       string
	ParentProductGroup string
	Storage            string
	Vendor             string
	ItemsPerUnit       float32
	QTY                float32
	SumQTY             float32
	Price2             float32
	RetailPrice        float32
	RetailCurrency     string
	CustPrice          float32
	Multiplicity       float32
	QTYlots            string
}

func GetItems(c *gin.Context) {

	db, err := sql.Open("odbc", "DSN=storage")
	if err != nil {
		fmt.Println("Error in connect DB")
		log.Fatal(err)
	}
	step, err := strconv.Atoi(c.DefaultQuery("step", "100"))
	if err != nil {
		step = 100
	}
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	if err != nil {
		start = 0
	}
	_, items := SelectRows(db, start, step)

	defer db.Close()

	c.IndentedJSON(http.StatusOK, items)
}

func GetItemCount(c *gin.Context) {

	db, err := sql.Open("odbc", "DSN=storage")
	if err != nil {
		fmt.Println("Error in connect DB")
		log.Fatal(err)
	}

	count, err := selectCountAll(db, "items")
	defer db.Close()

	c.IndentedJSON(http.StatusOK, count)

}

func selectCountAll(db *sql.DB, tableName string) (int, error) {

	var err error
	var count int
	var rows *sql.Rows

	query := "select count(*) from " + tableName
	rows, err = db.Query(query)

	if err != nil {
		log.Fatal(err)
	}
	rows.Next()
	if err = rows.Scan(&count); err != nil {
		log.Fatal(err)
	}
	//fmt.Println(count)

	defer rows.Close()

	return count, err
}

func SelectRows(db *sql.DB, start int, step int) (error, []Item) {
	var err error
	var query string
	var items []Item
	var item Item

	query = "SELECT items.id,brands.name brand,products.name productName,productGroups.name productGroup,parentProductGroups.name parentProductGroup," +
		"storages.name storage,vendors.name vendor,itemsPerUnit,QTY,sumQTY,price2,retailPrice,retailCurrency,custPrice,multiplicity,QTYlots FROM items " +
		"INNER JOIN brands ON(items.id_brand=brands.id) " +
		"INNER JOIN products ON(items.id_product=products.id) " +
		"INNER JOIN productGroups ON(items.id_productGroup=productGroups.id) " +
		"INNER JOIN parentProductGroups ON(items.id_parentProductGroup=parentProductGroups.id) " +
		"INNER JOIN storages ON(items.id_storage=storages.id) " +
		"INNER JOIN vendors ON(storages.id_vendor=vendors.id) " +
		"ORDER BY items.id " +
		"OFFSET " + strconv.Itoa(start) + " ROWS FETCH NEXT " + strconv.Itoa(step) + " ROWS ONLY"

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	//var byteValue []byte
	for rows.Next() {
		if err := rows.Scan(&item.Id, &item.Brand, &item.ProductName, &item.ProductGroup, &item.ParentProductGroup, &item.Storage, &item.Vendor, &item.ItemsPerUnit, &item.QTY, &item.SumQTY, &item.Price2, &item.RetailPrice, &item.RetailCurrency, &item.RetailPrice, &item.Multiplicity, &item.QTYlots); err != nil {
			log.Fatal(err)
		}

		err = decodeItem(&item)
		items = append(items, item)

	}

	defer rows.Close()
	return err, items
}

func decodeItem(item *Item) error {
	var err error

	item.Brand, _ = Decode(item.Brand)
	item.ProductName, _ = Decode(item.ProductName)
	item.ProductGroup, _ = Decode(item.ProductGroup)
	item.ParentProductGroup, _ = Decode(item.ParentProductGroup)
	item.Storage, _ = Decode(item.Storage)
	item.Vendor, _ = Decode(item.Vendor)

	return err
}

func Decode(document string) (string, error) {

	var err error

	decoder := charmap.Windows1251.NewDecoder()
	s, _ := decoder.String(document)

	return s, err
}

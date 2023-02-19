package dataBase

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/charmap"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type itemsByProvider struct {
	ProviderName string
	Items        []Item
}

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

type filter struct {
	Id   int
	Name string
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
	_, items := SelectRows(db, start, step, getFilters(c))

	fmt.Println(getFilters(c))

	defer db.Close()

	c.IndentedJSON(http.StatusOK, items)
}

func GetItemsByProviders(c *gin.Context) {

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
	filters := getFilterTable(db, "vendors")
	var resultItems []itemsByProvider
	var resultItem itemsByProvider

	for _, val := range filters {
		fmt.Println(val.Id)
		_, items := SelectRows(db, start, step, getFiltersByVendor(c, val.Id))
		resultItem.ProviderName = val.Name
		resultItem.Items = items
		resultItems = append(resultItems, resultItem)
	}

	defer db.Close()

	c.IndentedJSON(http.StatusOK, resultItems)
}

func getFilterTable(db *sql.DB, tableName string) []filter {
	var filterTable []filter
	var filterEx filter
	var err error
	var query string

	query = "SELECT * from " + tableName

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	//var byteValue []byte
	if tableName == "vendors" {
		var storageType string
		for rows.Next() {
			if err := rows.Scan(&filterEx.Id, &filterEx.Name, &storageType); err != nil {
				log.Fatal(err)
			}

			err = decodeFilter(&filterEx)
			filterTable = append(filterTable, filterEx)

		}
	}

	defer rows.Close()
	return filterTable

}

func getFilters(c *gin.Context) map[string]string {
	//var filters map[string]string
	filters := make(map[string]string)

	filters["products.name"] = c.Query("productName")
	filters["brands.id"] = c.Query("brandId")
	filters["productGroups.id"] = c.Query("productGroupId")
	filters["parentProductGroups.id"] = c.Query("parentProductGroupId")
	filters["storages.id"] = c.Query("storageId")
	filters["vendors.id"] = c.Query("vendorId")

	return filters
}

func getFiltersByVendor(c *gin.Context, providerId int) map[string]string {
	//var filters map[string]string
	filters := make(map[string]string)

	filters["products.name"] = c.Query("productName")
	filters["brands.id"] = c.Query("brandId")
	filters["productGroups.id"] = c.Query("productGroupId")
	filters["parentProductGroups.id"] = c.Query("parentProductGroupId")
	filters["storages.id"] = c.Query("storageId")
	filters["vendors.id"] = strconv.Itoa(providerId)

	return filters
}

func GetItemCount(c *gin.Context) {

	db, err := sql.Open("odbc", "DSN=storage")
	if err != nil {
		fmt.Println("Error in connect DB")
		log.Fatal(err)
	}

	count, err := selectItemsCount(db, getFilters(c))
	defer db.Close()

	c.IndentedJSON(http.StatusOK, count)

}

func selectItemsCount(db *sql.DB, filters map[string]string) (int, error) {

	var err error
	var count int
	var rows *sql.Rows

	query := "select count(*) from items " +
		"INNER JOIN brands ON(items.id_brand=brands.id) " +
		"INNER JOIN products ON(items.id_product=products.id) " +
		"INNER JOIN productGroups ON(items.id_productGroup=productGroups.id) " +
		"INNER JOIN parentProductGroups ON(items.id_parentProductGroup=parentProductGroups.id) " +
		"INNER JOIN storages ON(items.id_storage=storages.id) " +
		"INNER JOIN vendors ON(storages.id_vendor=vendors.id) " +
		"WHERE items.id = items.id "

	for idx, val := range filters {
		if idx == "products.name" && val != "" {
			names := strings.Split(val, " ")
			for _, v := range names {
				query = query + " AND " + idx + " LIKE " + "'%" + v + "%' "
			}

		} else if val != "" {
			query = query + " AND " + idx + " IN(" + val + ") "
		}
	}

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

func SelectRows(db *sql.DB, start int, step int, filters map[string]string) (error, []Item) {
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
		"WHERE items.id = items.id "

	for idx, val := range filters {
		if idx == "products.name" && val != "" {
			names := strings.Split(val, " ")
			for _, v := range names {
				query = query + " AND " + idx + " LIKE " + "'%" + v + "%' "
			}

		} else if val != "" {
			query = query + " AND " + idx + " IN(" + val + ") "

		}
	}
	query = query + "ORDER BY items.retailPrice OFFSET " + strconv.Itoa(start) + " ROWS FETCH NEXT " + strconv.Itoa(step) + " ROWS ONLY"

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

func decodeFilter(item *filter) error {
	var err error

	item.Name, _ = Decode(item.Name)

	return err
}

func Decode(document string) (string, error) {

	var err error

	decoder := charmap.Windows1251.NewDecoder()
	s, _ := decoder.String(document)

	return s, err
}

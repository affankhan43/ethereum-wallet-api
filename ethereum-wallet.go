package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"bytes"
	"reflect"
	//"context"
	"crypto/ecdsa"
	"database/sql"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Required struct {
	Coin   string   `json:"coin" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Key   string `json:"Key" binding:"required"`
}


func main() {
	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
	r := gin.Default()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"ethAPI": "password",
	}))
	authorized.POST("/getAddress", CreateAddress)

	r.Run(":8080")
}

func CreateAddress(c *gin.Context) {
	var required Required
	if err := c.BindJSON(&required); err != nil || required.Coin != "ETH" || required.Message != "chow_getaddress" || required.Key != "getaccess" {
		c.JSON(200, gin.H{
			"success": false,
			"message": "Access Denied",
		})
	} else {
		mysql_string:=os.Getenv("Mysql_access")+"@tcp("+os.Getenv("Mysql_link")+")/ethereum"
		fmt.Println(mysql_string)
		db, err := sql.Open("mysql", mysql_string)
		if err != nil {
			c.JSON(200,gin.H{
				"success":false,
				"message":"DB Connection Failed",
			})
		} else {
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				c.JSON(200,gin.H{
					"success":false,
					"message":"Address Generation Failed",
				})
			} else {
				privateKeyBytes := crypto.FromECDSA(privateKey)
				publicKey := privateKey.Public()
				publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
				if !ok {
					log.Fatal("error casting public key to ECDSA")
				}
				publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
				fmt.Println(hexutil.Encode(publicKeyBytes)[4:])
				pvKey:=hexutil.Encode(privateKeyBytes)
				pvEn := SplitSubN(pvKey,35)
				godotenv.Load()
				address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
				currentTime := time.Now()
				qrr, err := db.Query("INSERT INTO keystore (`label`,`address`,`key_data`,`key_next`,`created_at`) values ('test',?,hex(aes_encrypt(?,?)),hex(aes_encrypt(?,?)),?);",address,pvEn[0],os.Getenv("Secret_Key1"),pvEn[1],os.Getenv("Secret_Key2"),currentTime.Format("2006.01.02 15:04:05"))
				if err != nil {
					fmt.Println(err)
					fmt.Println(qrr)
					c.JSON(200,gin.H{
						"success":false,
						"message":"DB Insert Failed",
					})
				} else {
					c.JSON(200,gin.H{
						"success":true,
						"address":address,
						"pv":pvKey,
					})
				}
			}
		}
	}
}

func in_array(val interface{}, array interface{}) (exists bool) {
	exists = false
	//index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				// index = i
				exists = true
				return
			}
		}
	}
	return
}

func SplitSubN(s string, n int) []string {
	sub := ""
	subs := []string{}

	runes := bytes.Runes([]byte(s))
	l := len(runes)
	for i, r := range runes {
		sub = sub + string(r)
		if (i + 1) % n == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}
	return subs
}
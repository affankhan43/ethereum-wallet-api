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

type Check struct {
	Auth Required `json:"auth" binding:"required"`
	Previous int `json:"previous" binding:"required"`
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

func CheckDeposits(c *gin.Context) {
	var required Check
	if err := c.BindJSON(&required); err != nil || required.Auth.Coin != "ETH" || required.Auth.Message != "Check_deposita" || required.Auth.Key != "getaccess" {
		c.JSON(200, gin.H{
			"success": false,
			"message": "Access Denied",
		})
	} else{
		mysql_string:=os.Getenv("Mysql_access")+"@tcp("+os.Getenv("Mysql_link")+")/ethereum"
		db, err := sql.Open("mysql", mysql_string)
		if err != nil {
			c.JSON(200,gin.H{
				"success":false,
				"message":"DB Connection Failed",
			})
		} else{
			results, err := db.Query("SELECT id,address FROM keystore")
			if err != nil{
				c.JSON(200,gin.H{
					"success":false,
					"message":"DB Connection Failed",
				})
			} else {
				addresses:= map[string]int{}
				for results.Next() {
					var address string
					var id int
					err = results.Scan(&id,&address)
					if err != nil {
						fmt.Println(err)
					}
					addresses[address] = id
				}
				client, err := ethclient.Dial("https://mainnet.infura.io")
				if err != nil {
					c.JSON(200,gin.H{
						"success":false,
						"message":"Blockchain Connection Failed",
					})
				} else {
					header, err := client.HeaderByNumber(context.Background(), nil)
					if err != nil {
						c.JSON(200,gin.H{
							"success":false,
							"message":"Blockchain Connection Failed",
						})
					} else{
						latest:=header.Number.Int64()-11
						checkno:=latest-int64(required.Previous)
						previous:= int64(required.Previous)
						var transactions = map[int]map[string]string{}
						txno:=0
						if checkno <= 0 {
							c.JSON(200,gin.H{
								"success":true,
								"transactions":transactions,
								"latest":latest,
							})
						} else {
							for i:=int64(0); i<checkno; i++ {
								blockNumber := big.NewInt(previous)
								fmt.Println(previous)
								block, err := client.BlockByNumber(context.Background(), blockNumber)
								if err != nil {
                           c.JSON(200,gin.H{
                              "success":false,
                              "message":"Blockchain Connection Failed",
                              })
                        } else{
                           for _, tx := range block.Transactions() {
                              if tx.To() != nil{
                                 if aid, founds := addresses[tx.To().String()]; founds {
                                    receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
                                    if err != nil {
                                       c.JSON(200,gin.H{
                                          "success":false,
                                          "message":"Blockchain Connection Failed",
                                          })
                                    }
                                    //fmt.Println(receipt.Status)
                                    if receipt.Status == 1 {
                                       fmt.Println(aid)
                                       conf:=header.Number.Int64()-int64(previous)
                                       sconf:=strconv.FormatInt(conf, 10)
                                       value := new(big.Float)
                                       value.SetString(tx.Value().String())
                                       ethValue := new(big.Float).Quo(value, big.NewFloat(math.Pow10(18)))
                                       transactions[txno] = map[string]string{}
                                       transactions[txno]["coin"]="ETH"
                                       transactions[txno]["txid"]=tx.Hash().String()
                                       transactions[txno]["to"]=tx.To().String()
                                       transactions[txno]["value"]=ethValue.String()
                                       transactions[txno]["confirmation"]=sconf
                                       txno+=1
                                    }
                                 } else if v, found := erc20s[tx.To().String()]; found {
                                    fmt.Println(v)
                                    receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
                                    if err != nil {
                                       fmt.Println(err)
                                    }
                                    if receipt.Status == 1 {
                                       if len(receipt.Logs) > 0 && len(receipt.Logs[0].Topics) == 3 {
                                          // jsonString,_:=json.Marshal(receipt.Logs[0])
                                          // fmt.Println(string(jsonString))
                                          to:=common.BytesToAddress(receipt.Logs[0].Topics[2].Bytes())
                                          if aids, fds := addresses[to.String()]; fds {
                                             fmt.Println(aids)
                                             transactions[txno] = map[string]string{}
                                             //from:=common.BytesToAddress(receipt.Logs[0].Topics[1].Bytes())
                                             b := big.NewInt(0)
                                             val:=b.SetBytes(receipt.Logs[0].Data)
                                             conf:=header.Number.Int64()-int64(previous)
                                             sconf:=strconv.FormatInt(conf, 10)
                                             value := new(big.Float)
                                             value.SetString(val.String())
                                             powd,_:= strconv.Atoi(v["dec"])
                                             ethValue := new(big.Float).Quo(value, big.NewFloat(math.Pow10(powd)))
                                             transactions[txno]["coin"]=v["token"]
                                             transactions[txno]["txid"]=tx.Hash().String()
                                             transactions[txno]["to"]=to.String()
                                             transactions[txno]["value"]=ethValue.String()
                                             transactions[txno]["confirmation"]=sconf
                                             txno+=1
                                          }
                                       }
                                    }
                                 }
                              }
                           }
                           previous+=int64(1)
                        }
                     }
                     c.JSON(200, gin.H{
                        "success":true,
                        "transactions":transactions,
                        "block":latest,
                        })
                  }
               }
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
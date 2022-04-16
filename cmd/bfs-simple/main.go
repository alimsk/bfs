package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"time"

	"github.com/alimsk/shopee"
	jsoniter "github.com/json-iterator/go"
)

var version string

func main() {
	flag.Parse()
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	if flag.NArg() > 0 {
		switch flag.Arg(0) {
		case "login":
		}
		return
	}

	b, err := io.ReadAll(os.Stdin)
	fatalfIf("error reading stdin: %v", err)
	b = bytes.TrimSpace(b)

	c, err := loginFromCookieJson(b)
	if err != nil {
		c, err = shopee.NewFromCookieString(string(b))
		if err != nil {
			log.Fatal(err)
		}
	}

	addrs, err := c.FetchAddresses()
	fatalIf(err)
	i, addr := addrs.DeliveryAddress()
	if i == -1 {
		log.Fatal("alamat utama tidak disetting, silahkan setting terlebih dahulu")
	}

	urlstr := input("URL: ")
	item, err := c.FetchItemFromURL(urlstr)
	fatalIf(err)
	fmt.Println(item.Name())

	fmt.Println("\nPilih Model")
	for i, m := range item.Models() {
		fmt.Println()
		fmt.Println(i, m.Name())
		fmt.Println("id:", m.ModelID())
		fmt.Println("stok:", m.Stock())
		fmt.Println("harga:", m.Price())
		fmt.Println("flashsale:", m.HasUpcomingFsale())
	}
	model := item.Models()[inputint("Pilih:")]

	fmt.Println("\nMetode Pembayaran")
	for i, ch := range shopee.PaymentChannelList {
		fmt.Println(i, ch.Name)
	}
	paymentch := shopee.PaymentChannelList[inputint("Pilih: ")]
	var paymentdata shopee.PaymentChannelData
	if len(paymentch.Options) > 0 {
		for i, ch := range paymentch.Options {
			fmt.Println(i, ch.Name)
		}
		paymentdata = paymentch.ApplyOpt(paymentch.Options[inputint("Pilih:")])
	} else {
		paymentdata = paymentch.Apply()
	}

	fmt.Println("\nmengambil info logistik")
	logistics, err := c.FetchShippingInfo(addr, item)
	fatalIf(err)

	{
		tmp := logistics[:0]
		for _, logistic := range logistics {
			if !logistic.HasWarning() {
				tmp = append(tmp, logistic)
			}
		}
		logistics = tmp
	}

	if len(logistics) == 0 {
		log.Fatal("tidak ada channel logistik yang tersedia")
	}

	fmt.Println("\nChannel Logistik")
	for i, logistic := range logistics {
		fmt.Println(i, logistic.Name())
	}
	logistic := logistics[inputint("Pilih: ")]

	citem := shopee.ChooseModel(item, model.ModelID())
	fstime := time.Unix(item.UpcomingFsaleStartTime(), 0)
	log.Println("flash sale pada", fstime.Format("3:04:05 PM"))
	log.Println("start validasi checkout")
	fatalIf(c.ValidateCheckout(citem))
	log.Println("validasi checkout done")
	log.Println("start checkout get")
	params := shopee.CheckoutParams{
		Addr:        addr,
		Item:        citem,
		PaymentData: paymentdata,
		Logistic:    logistic,
	}.WithTimestamp(time.Now().Unix())
	_, err = c.CheckoutGetQuick(params)
	fatalIf(err)
	log.Println("checkout get done")
	log.Println("start place order")
	fatalIf(c.PlaceOrder(params))
	log.Println("place order done")
}

func loginFromCookieJson(b []byte) (shopee.Client, error) {
	if !jsoniter.Valid(b) {
		return shopee.Client{}, errors.New("not a valid json input")
	}

	json := jsoniter.Get(b)
	cookies := make([]*http.Cookie, json.Size())
	for i := 0; i < json.Size(); i++ {
		item := json.Get(i)
		value, err := strconv.Unquote(item.Get("value").ToString())
		if err != nil {
			value = item.Get("value").ToString()
		}
		// do not set expires
		cookies[i] = &http.Cookie{
			Name:   item.Get("name").ToString(),
			Value:  value,
			Domain: item.Get("domain").ToString(),
			// Expires:  time.Unix(item.Get("expirationDate").ToInt64(), 0),
			HttpOnly: item.Get("httpOnly").ToBool(),
			Path:     item.Get("path").ToString(),
			Secure:   item.Get("secure").ToBool(),
		}
	}

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(shopee.ShopeeUrl, cookies)
	return shopee.New(jar)
}

func input(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		log.Fatal(scanner.Err())
	}
	return scanner.Text()
}

func inputint(prompt string) int {
	for {
		inp := input(prompt)
		if v, err := strconv.Atoi(inp); err == nil {
			return v
		}
		fmt.Println("masukkan angka")
	}
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func fatalfIf(format string, err error) {
	if err != nil {
		log.Fatalf(format, err)
	}
}

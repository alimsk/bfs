package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alimsk/shopee"
	jsoniter "github.com/json-iterator/go"
)

var version string

var (
	delay      = flag.Duration("d", 0, "delay antar request saat checkout")
	subFSTime  = flag.Duration("sub", 0, "kurangi waktu flash sale")
	cookieFile = flag.String("f", "cookie", "cookie file")
)

// https://github.com/golang/go/issues/20455#issuecomment-342287698
func fixTimezone() {
	out, err := exec.Command("/system/bin/getprop", "persist.sys.timezone").Output()
	if err != nil {
		return
	}
	z, err := time.LoadLocation(strings.TrimSpace(string(out)))
	if err != nil {
		return
	}
	time.Local = z
}

func init() {
	if runtime.GOOS == "android" {
		fixTimezone()
	}
}

func cookiesToString(cookies []*http.Cookie) []byte {
	r, _ := http.NewRequest("", "", nil)
	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}
	return []byte(r.Header.Get("Cookie"))
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	if flag.NArg() > 0 {
		switch flag.Arg(0) {
		case "info":
			itemInfo()
		case "version":
			fmt.Println(version, "github.com/alimsk/bfs")
		default:
			log.Fatal("unknown subcommand: ", flag.Arg(0))
		}
		return
	}

	f, err := os.Open(*cookieFile)
	fatalIf(err)

	b, err := io.ReadAll(f)
	fatalfIf("error reading stdin: %v", err)
	b = bytes.TrimSpace(b)
	f.Close()

	log.Println("attempting to parse cookie json")
	c, err := loginFromCookieJson(b)
	if err != nil {
		log.Println("error:", err)
		log.Println("parsing cookie string")
		c, err = shopee.NewFromCookieString(string(b))
		if err != nil {
			log.Fatal(err)
		}
	}

	acc, err := c.FetchAccountInfo()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		ioutil.WriteFile(*cookieFile, cookiesToString(c.Client.GetClient().Jar.Cookies(shopee.ShopeeUrl)), 0644)
	}()
	fmt.Println("login sebagai", acc.Username())

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
		fmt.Println("harga:", formatPrice(m.Price()))
		fmt.Println("flashsale mendatang:", m.HasUpcomingFsale())
	}
	fmt.Println()
	model := item.Models()[inputint("Pilih: ")]

	fmt.Println("\nMetode Pembayaran")
	PaymentChannelList := [...]shopee.PaymentChannel{shopee.ShopeePay, shopee.COD, shopee.TransferBank, shopee.Alfamart, shopee.Indomaret}
	for i, ch := range PaymentChannelList {
		fmt.Println(i, ch.Name)
	}
	fmt.Println()
	paymentch := PaymentChannelList[inputint("Pilih: ")]
	var paymentOption string
	if len(paymentch.Options()) > 0 {
		for i, ch := range paymentch.Options() {
			fmt.Println(i, ch.Name)
		}
		fmt.Println()
		paymentOption = paymentch.Options()[inputint("Pilih:")].OptionInfo
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
		fmt.Println(i, logistic.Name(), "|", formatPrice(logistic.PriceBeforeDiscount()))
	}
	fmt.Println()
	logistic := logistics[inputint("Pilih: ")]

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	fstime := time.Unix(item.UpcomingFsaleStartTime(), 0)
	citem := shopee.ChooseModel(item, model.ModelID())

	if fstime.After(time.Now()) {
		log.Println("flash sale pada", fstime.Format("3:04:05 PM"))
		time.Sleep(time.Until(fstime) - *subFSTime)

		log.Println("start refresh item")
		item, err = c.FetchItem(item.ShopID(), item.ItemID())
		fatalIf(err)
		log.Println("finish refresh item")
		citem = shopee.ChooseModel(item, model.ModelID())
	}

	params := shopee.CheckoutParams{
		Addr:          addr,
		Item:          citem,
		Payment:       paymentch,
		PaymentOption: paymentOption,
		Logistic:      logistic,
	}

	if *delay == 0 {
		nodelay(c, citem, params)
		return
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		log.Println("start validasi checkout")
		fatalIf(c.ValidateCheckout(citem))
		log.Println("finish validasi checkout")
		wg.Done()
	}()
	time.Sleep(*delay)
	params = params.WithTimestamp(time.Now().Unix())
	go func() {
		log.Println("start checkout get")
		_, err = c.CheckoutGetQuick(params)
		fatalIf(err)
		log.Println("finish checkout get")
		wg.Done()
	}()
	time.Sleep(*delay)
	go func() {
		log.Println("start place order")
		fatalIf(c.PlaceOrder(params))
		log.Println("finish place order")
		wg.Done()
	}()
	wg.Wait()
}

func nodelay(c shopee.Client, citem shopee.CheckoutableItem, params shopee.CheckoutParams) {
	log.Println("start validasi checkout")
	fatalIf(c.ValidateCheckout(citem))
	log.Println("finish validasi checkout")
	log.Println("start checkout get")
	params, err := c.CheckoutGetQuick(params)
	fatalIf(err)
	log.Println("finish checkout get")
	log.Println("start place order")
	fatalIf(c.PlaceOrder(params))
	log.Println("finish place order")
}

func itemInfo() {
	urlstr := flag.Arg(1)

	c, err := shopee.NewFromCookieString("csrftoken=" + randstr(32))
	if err != nil {
		log.Fatal(err)
	}

	item, err := c.FetchItemFromURL(urlstr)
	if err != nil {
		log.Fatal(err)
	}

	fsalestatus := "tidak ada"
	if item.IsFlashSale() {
		fsalestatus = "sedang berlangsung"
	} else if item.HasUpcomingFsale() {
		fsalestatus = "pada jam " + time.Unix(item.UpcomingFsaleStartTime(), 0).Format("3:04:05 PM")
	}

	m := [...]struct {
		k string
		v interface{}
	}{
		{"Flashsale", fsalestatus},
		{"Harga", formatPrice(item.Price())},
		{"Stok", item.Stock()},
		{"Kategori", strings.Join(item.CatNames(), ", ")},
		{"Shopid", item.ShopID()},
		{"Itemid", item.ItemID()},
	}

	var longestkey int
	for _, v := range m {
		if len(v.k) > longestkey {
			longestkey = len(v.k)
		}
	}

	fmt.Println(item.Name())
	fmt.Println()
	for _, v := range m {
		fmt.Printf("%-*s %v\n", longestkey+1, v.k+":", v.v)
	}

	for _, model := range item.Models() {
		fmt.Println(
			"\n"+model.Name(),
			"\nID:                 ", model.ModelID(),
			"\nHarga:              ", formatPrice(model.Price()),
			"\nStok:               ", model.Stock(),
			"\nFlashsale Mendatang:", ternary(model.HasUpcomingFsale(), "Ya", "Tidak"),
		)
	}
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

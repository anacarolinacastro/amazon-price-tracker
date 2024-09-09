package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func getBookHTMLPage(isbn string) *html.Node {
	bookURL, _ := url.Parse(fmt.Sprintf("https://www.amazon.com.br/dp/%s", isbn))

	expiration := time.Now().Add(5 * time.Minute)
	cookies := []*http.Cookie{
		&http.Cookie{Name: "csm-sid", Value: "232-3081956-9165437", Expires: expiration},
		&http.Cookie{Name: "x-amz-captcha-1", Value: "1725839342323476", Expires: expiration},
		&http.Cookie{Name: "x-amz-captcha-2", Value: "gAwxPrYxYRq6ogiyep5IQA==", Expires: expiration},
		&http.Cookie{Name: "session-id", Value: "147-8639576-8381120", Expires: expiration},
		&http.Cookie{Name: "i18n-prefs", Value: "BRL", Expires: expiration},
		&http.Cookie{Name: "ubid-acbbr", Value: "133-2159547-1471115", Expires: expiration},
		&http.Cookie{Name: "csm-hit", Value: "tb:s-EVPDJ8DN7PSCBXMJSWKQ|1725832142453&t:1725832143355&adb:adblk_yes", Expires: expiration},
		&http.Cookie{Name: "session-token", Value: `isBpNlvWjgv1Ow3CVvpC2EjDipBiuidJiAqvVcNCu6BDIaiDkvkmmvmEFgJZThYo8HnjB0I48V6wgax/z2ldfOIiNPO6xd885IHEs0vYGJt8E+EbjjK2HouBTmjDAvFFZyC0Zo+rMFpQ1uaL3vLsH9nz+1EPMWKZps7vCMHNrWH4rQYo93rUz7s9tK5W7ASrBuIzmj7piOT1gyeedDkwkkEXEKJ+AhVIk2AV2WaGYbAh1SmTmZhdXVHZBAB8+iw7X6V+k9MOX8BgY96S2hZ7r332HBdtx95rxTxmV9y4qknaycS4nkFLcBZX88Fovf5Uk9be1t+jnjrcqVf3gzGPuBh54ppq+rMuQeUsWQor/8c=`, Expires: expiration},
		&http.Cookie{Name: "session-id-time", Value: "2082787201l'", Expires: expiration},
	}

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(bookURL, cookies)

	client := &http.Client{
		Jar: jar,
	}

	req, err := http.NewRequest("GET", bookURL.String(), nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:50.0) Gecko/20100101 Firefox/50.0")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode > 399 {
		fmt.Printf("Error fetching book: %d\n", isbn)
		panic(err)
	}

	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("Error finding book: %d\n", isbn)
		panic(err)
	}

	return doc
}

type ProductData struct {
	ISBN     string
	Title    string
	Reais    int
	Centavos int
}

func processPrice(n *html.Node, productData *ProductData) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		for _, a := range c.Attr {
			if a.Key == "class" && strings.Contains(a.Val, "a-price-whole") {
				reais, _ := strconv.Atoi(c.FirstChild.Data)
				productData.Reais = reais
			}

			if a.Key == "class" && strings.Contains(a.Val, "a-price-fraction") {
				centavos, _ := strconv.Atoi(c.FirstChild.Data)
				productData.Centavos = centavos
			}
		}
	}
}

func processNode(n *html.Node, productData *ProductData) {
	switch n.Data {
	case "span":
		for _, a := range n.Attr {
			if a.Key == "id" && strings.Contains(a.Val, "productTitle") {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						productData.Title = c.Data
					}
				}
			}

			if a.Key == "class" && strings.Contains(a.Val, "a-price aok-align-center reinventPricePriceToPayMargin priceToPay") {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Attr[0].Key == "aria-hidden" {
						processPrice(c, productData)
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNode(c, productData)
	}
}

func main() {
	ISBNList := []string{
		"6560050548",
		"8595085927",
		"8595086788",
		"6555115718",
		"6560051099",
		"6555111755",
		"6555112247",
		"8595085730",
		"6555110023",
		"6560051501",
		"6555113561",
		"8595085900",
		"6560050718",
		"655511309X",
		"6555111089",
		"6555111348",
		"859508548X",
		"8595085943",
		"6555111984",
		"8595086796",
		"859508680X",
		"6555112859",
		"6555111992",
		"6555112239",
		"6555112506",
		"8595085919",
		"6560050114",
		"6555115130",
		"6555114703",
		"8595085935",
		"6560050548",
		"6555111364",
		"6555111097",
		"8595085935",
		"6555114444",
		"6555113901",
	}
	productsData := []ProductData{}

	for _, isbn := range ISBNList {
		doc := getBookHTMLPage(isbn)
		bookData := ProductData{ISBN: isbn}
		processNode(doc, &bookData)
		fmt.Printf("Book: %v\n", bookData)
		productsData = append(productsData, bookData)
		time.Sleep(15 * time.Second)
	}
}

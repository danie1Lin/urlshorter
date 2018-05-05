package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"time"
)

type MyMux struct {
	Controllers map[string]*Controller
}

type Controller struct {
	Data *mgo.Collection
}

func (c *Controller) GET(w http.ResponseWriter, r *http.Request) (err error) {
	//handle data
	//Render

	result := make(Urlpairs, 0)
	c.Data.Find(bson.M{"short": bson.M{"$regex": ".*"}}).All(&result)
	t, _ := template.ParseFiles("shortList.gtpl")
	t.Execute(w, result)
	return nil
}

func (c *Controller) Redirect(w http.ResponseWriter, r *http.Request, short string) {
	fmt.Println("Redirect")
	result := Urlpair{}
	c.Data.Find(bson.M{"short": short}).One(&result)
	fmt.Println("before: ", result)
	time.Sleep(1000)
	result.Clicked += 1
	c.Data.Update(bson.M{"short": short}, result)
	fmt.Println("After: ", result)
	http.Redirect(w, r, result.OriginalUrl, 301)
}

func (c *Controller) POST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.Form.Get("url")
	_ = AddUrl(c.Data, url)
}

type MgoError struct {
	Op  string
	Err error
}

func (e *MgoError) Error() string {
	return e.Op + ":" + e.Err.Error()
}

func (m *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	split := regexp.MustCompile("/")
	path := split.Split(r.URL.Path[1:], -1)
	if path[len(path)-1] == "" {
		path = path[:len(path)-1]
	}
	fmt.Println(path)

	if path[0] == "short" {
		switch a := len(path); {
		case a == 1:
			//show shortUrllist
			if r.Method == "GET" {
				m.Controllers["shorter"].GET(w, r)
			} else {
				m.Controllers["shorter"].POST(w, r)
				m.Controllers["shorter"].GET(w, r)
			}
		case a > 1:
			//Redirect
			fmt.Println("Porting...")
			m.Controllers["shorter"].Redirect(w, r, path[1])
		//decode

		//redirect
		default:
			fmt.Println("Url not found")
		}
	}

	return
}

func short(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("shortList.gtpl")
	t.Execute(w, nil)
}

func gifurl(w http.ResponseWriter, r *http.Request) {

}

type Urlpair struct {
	OriginalUrl string
	Short       string
	Clicked     int
}

type Urlpairs []Urlpair

var Base62 string = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`

func encode(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	token := fmt.Sprintf("%x", h.Sum(nil))
	return token
}

func idexBase62(idx int) string {
	a := float64(idx)
	base := float64(len(Base62))
	r := ""
	for i := 0; base < a || i < 2; i++ {
		a = math.Floor(a / base)
		var c int
		if a == 0 {
			c = int(a)
			a = float64(idx)
		} else {
			c = int(math.Mod(a, base))
		}
		r += string(Base62[c])
		fmt.Println("for:", r, a, c)
	}
	c := int(a)
	r += string(Base62[c])
	fmt.Print("for:", r, a, c)
	return r
}

func MgoInit() *mgo.Collection {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			fmt.Printf("%T \n\r", err)
			a := &MgoError{"fuck", errors.New("MgoInit")}
			fmt.Println(a.Error())
			//panic("Please open mangodb server")
		}
	}()
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		fmt.Println("mongodb connecting error :", err)
	}
	c := session.DB("test").C("short")
	fmt.Println(c.Count())
	result := make(Urlpairs, 0)
	c.Find(bson.M{"short": "aaa"}).All(&result)
	fmt.Println("bson", result)
	_ = AddUrl(c, "https://astaxie.gitbooks.io/build-web-application-with-golang/content/zh/05.6.html")
	return c
}

func AddUrl(c *mgo.Collection, url string) string {
	//check if there is same original url
	var encoded string
	n, _ := c.Find(bson.M{"originalurl": url}).Count()
	if n == 0 {
		idx, _ := c.Count()
		encoded = idexBase62(idx)
		fmt.Println("New : ", encoded)
	} else {
		result := Urlpair{}
		c.Find(bson.M{"originalurl": url}).One(&result)
		fmt.Println("existed:", result)
		return result.Short
	}
	//check shorturl unique
	c.Insert(Urlpair{url, encoded, 0})
	result := make(Urlpairs, 0)
	c.Find(bson.M{"short": encoded}).All(&result)
	fmt.Println(result)
	return encoded
	//if not create a new one
}

func main() {
	//http.HandleFunc("/short", short)

	c := MgoInit()
	controller := Controller{c}
	mux := &MyMux{map[string]*Controller{"shorter": &controller}}
	err := http.ListenAndServe(":8000", mux)
	if err != nil {
		log.Fatal(err)
	}
}

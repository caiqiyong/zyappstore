package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

type AppInfo struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	AppVersion string `json:"appVersion"`
	AppSize    string `json:"appSize"`
	AppName    string `json:"appName"`
	AppDesc    string `json:"appDesc"`
	AppUrl     string `json:"appUrl"`
	Explain    string `json:"explain"`
	CategoryId int    `json:"categoryId"`
}

type PageInfo struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalPage   int `json:"totalPage"`
	TotalRecord int `json:"totalRecord"`
}

var gAppList map[string]AppInfo
var gMutex sync.Mutex
var gFileSize int64
var gFileMTime time.Time

var gIp = flag.String("ip", "", "ip")
var gPort = flag.Int("port", 8889, "port")
var gRuning = true

func replyError(w http.ResponseWriter, msg string) {
	res := map[string]interface{}{}
	res["code"] = -1
	res["msg"] = msg
	b, _ := json.Marshal(res)
	w.Write(b)
}

func handleGetAppList(w http.ResponseWriter, params url.Values) {
	value, ok := params["page"]
	if !ok {
		replyError(w, "miss field page")
		return
	}
	nPage, err := strconv.ParseInt(value[0], 10, 0)
	if err != nil {
		replyError(w, "invalid feild page")
		return
	}

	value, ok = params["size"]
	if !ok {
		replyError(w, "miss field size")
		return
	}
	nSize, err := strconv.ParseInt(value[0], 10, 0)
	if err != nil {
		replyError(w, "invaild field size")
		return
	}

	res := map[string]interface{}{}
	res["code"] = 0
	res["msg"] = ""

	list := make([]AppInfo, 0, 1000)
	gMutex.Lock()
	for _, app := range gAppList {
		list = append(list, app)
	}
	gMutex.Unlock()

	page := []AppInfo{}
	l := len(list)
	totalPage := 0
	m := l % int(nSize)
	if m > 0 {
		totalPage = l/int(nSize) + 1
	} else {
		totalPage = l / int(nSize)
	}

	if int64(l) >= nPage*nSize {
		page = list[((nPage - 1) * nSize):(nPage * nSize)]
	} else if int64(l) > (nPage-1)*nSize {
		page = list[((nPage - 1) * nSize):l]
	}

	p := PageInfo{
		CurrentPage: int(nPage),
		PageSize:    int(nSize),
		TotalPage:   totalPage,
		TotalRecord: l,
	}

	body := map[string]interface{}{}
	body["page"] = p
	body["list"] = page

	res["body"] = body

	buf, err := json.Marshal(res)
	if err != nil {
		replyError(w, "server exception")
		return
	}

	w.Write(buf)
}

func handleGetAppInfo(w http.ResponseWriter, params url.Values) {
	appName, ok := params["appName"]
	if !ok {
		replyError(w, "miss field appName")
		return
	}

	res := map[string]interface{}{}
	res["code"] = 0
	res["msg"] = ""

	app := AppInfo{}
	gMutex.Lock()
	app, _ = gAppList[appName[0]]
	gMutex.Unlock()

	if ok {
		res["body"] = app
	}

	buf, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("json.Marshall err %v\n", err)
		return
	}

	w.Write(buf)
}

func handleListApps(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	req := r.URL.RequestURI()
	fmt.Printf("%s\n", req)

	t, ok := params["ca"]
	if !ok {
		replyError(w, "miss field ca")
		return
	}

	switch t[0] {
	case "Eink_AppStore.AppList":
		handleGetAppList(w, params)
	case "Eink_AppStore.AppInfo":
		handleGetAppInfo(w, params)
	default:
		replyError(w, "invalid ca field")
	}
}

func isAppVaild(app AppInfo) bool {
	return app.AppName != "" && app.AppUrl != "" && app.Name != ""
}

func loadApps() bool {

	data, err := os.ReadFile("applist.json")
	if err != nil {
		fmt.Printf("open file error %v", err)
		return false
	}
	res := []AppInfo{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		fmt.Printf("json.Unmarshal() err %v\n", err)
		return false
	}

	gMutex.Lock()
	gAppList = make(map[string]AppInfo, 1000)
	for _, app := range res {
		if !isAppVaild(app) {
			continue
		}
		gAppList[app.AppName] = app
	}
	gMutex.Unlock()

	return true
}

func main() {

	flag.Parse()

	fmt.Printf("server linsten on %s:%d\n", *gIp, *gPort)
	server := fmt.Sprintf("%s:%d", *gIp, *gPort)

	http.HandleFunc("/zybook3/app/app.php", handleListApps)
	http.HandleFunc("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("app"))).ServeHTTP)

	go func() {
		err := http.ListenAndServe(server, nil)
		if err != nil {
			fmt.Printf("server start failed %v\n", err)
		}

		gRuning = false
	}()

	for gRuning {
		f, err := os.Stat("applist.json")
		if err != nil {
			fmt.Printf("os.Stat() error %v\n", err)
			break
		}
		if f.ModTime() != gFileMTime || f.Size() != gFileSize {
			if !loadApps() {
				break
			}
			gFileMTime = f.ModTime()
			gFileSize = f.Size()
		}
		time.Sleep(2 * time.Second)
	}
}

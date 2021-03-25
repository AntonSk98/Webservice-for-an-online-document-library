package main

import (
    "fmt"
    "github.com/go-zookeeper/zk"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "time"
)

const iconsCssPath = "/css/icons.css"
const materializeCssPath = "/css/materialize.min.css"
const jqueryPath = "/js/jquery-2.2.4.min.js"
const materializeJsPath = "/js/materialize.min.js"
const libraryPrefix = "/library"
const ngINX = "http://nginx"

var gServer = "http://gserve1"
var roundRobinServerCount = 0
func main() {
    mux := http.NewServeMux()

    connection, _, _ := zk.Connect([]string{"zookeeper:2181"}, time.Second)
    proxy := &httputil.ReverseProxy{}
    //proxyGServer := &httputil.ReverseProxy{Director: gServeDirector}

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if strings.Contains(r.URL.Path, libraryPrefix) {
            gServer = chooseServerToHandleUrl(connection)
            fmt.Println("SERVER PATH TO HANDLE: "+ gServer)
            proxy.Director = gServeDirector
            proxy.ServeHTTP(w,r)
        } else {
            proxy.Director = ngInxDirector
            proxy.ServeHTTP(w,r)
        }
    })
    log.Fatal(http.ListenAndServe(":80", mux))
}

func ngInxDirector (req *http.Request) {
    ngInxUrl, _ := url.Parse(ngINX)
    req.Header.Add("X-Forwarded-Host", req.Host)
    req.Header.Add("X-Origin-Host", ngInxUrl.Host)
    req.Host = ngInxUrl.Host
    req.URL.Scheme = "http"
    req.URL.Host = ngInxUrl.Host
    req.URL.Path = handleUrlForPath(req.URL.Path)
    fmt.Println("Redirect request to nginx....")
}

func gServeDirector (req *http.Request) {
    gServerUrl, _ := url.Parse(gServer)
    req.Header.Add("X-Forwarded-Host", req.Host)
    req.Header.Add("X-Origin-Host", gServerUrl.Host)
    req.Host = gServerUrl.Host
    req.URL.Scheme = "http"
    req.URL.Host = gServerUrl.Host
    fmt.Println("Redirect request to gserve...")
}

func handleUrlForPath(url string) string {
    if strings.Contains(url, iconsCssPath) {
        return iconsCssPath
    } else if strings.Contains(url, materializeCssPath){
        return materializeCssPath
    } else if strings.Contains(url, jqueryPath) {
        return jqueryPath
    } else if strings.Contains(url, materializeJsPath){
        return materializeJsPath
    } else {
        return "/"
    }
}

func chooseServerToHandleUrl(connection *zk.Conn) string {
    pathUrl := "http://"
    runningServices, _, err := connection.Children("/services")
    if err != nil {
        log.Fatal(err)
    }
    if roundRobinServerCount >= len(runningServices) {
        roundRobinServerCount = 0
    }
    pathUrl = pathUrl+runningServices[roundRobinServerCount]
    roundRobinServerCount++
    return pathUrl
}
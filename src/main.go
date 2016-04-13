package main

import (
	"apis"
	"config"
	"express"
	"logger"
	"loop"
	"net/http"
	"tools"
)

func main() {
	defer logger.Flush()
	logger.SetLevel("DEBUG")
	// 启动主循环
	go loop.Loop()
	router := express.NewRouter()
	// 添加过滤器
	router.Filter(tools.FilterLogging)
	router.Filter(tools.FilterAuthUser)
	// 添加路由
	router.HandleFunc("GET", "/urls", apis.UrlsDir)
	router.HandleFunc("PUT", "/urls", apis.UrlsAdd)
	router.HandleFunc("GET", "/urls/{UrlId}", apis.UrlsInfo)
	router.HandleFunc("POST", "/urls/{UrlId}", apis.UrlsUpdate)
	router.HandleFunc("DELETE", "/urls/{UrlId}", apis.UrlsDelete)

	router.HandleFunc("GET", "/users/self", apis.UsersInfo)
	router.HandleFunc("POST", "/users/self", apis.UsersUpdate)
	router.HandleFunc("GET", "/users/qrcode", apis.UsersQRCode)

	router.HandleFunc("GET", "/users/member", apis.UsersMemerList)
	router.HandleFunc("PUT", "/users/member", apis.UsersMemberAdd)
	router.HandleFunc("DELETE", "/users/member", apis.UsersMemberDel)

	router.HandleFunc("GET", "/users/group", apis.UsersGroupList)
	router.HandleFunc("DELETE", "/users/group", apis.UsersGroupDel)

	router.HandleFunc("GET", "/weixin", apis.WeiXinGet)
	router.HandleFunc("POST", "/weixin", apis.WeiXinPost)
	router.HandleFunc("GET", "/weixin/oauth", apis.WeiXinOAuth)

	router.HandleFunc("GET", "/queue/join", apis.QueueJoin)
	// 默认路由
	router.DefaultHandle(defaultHandler)
	logger.Infof("Starting listen on %s", config.Listener)
	logger.Fatal(http.ListenAndServe(config.Listener, router))
}

func defaultHandler(w *express.Response, r *express.Request) {
	w.Status(http.StatusNotFound).Json(apis.NewJSONError("404 Resource not found"))
}

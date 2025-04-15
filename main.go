// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(ctx *fasthttp.RequestCtx) {
	log.Println(string(ctx.Path()))

	if !ctx.IsGet() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}
	fasthttp.ServeFile(ctx, "./home.html")
}

func js(ctx *fasthttp.RequestCtx) {
	log.Println(string(ctx.Path()))

	if !ctx.IsGet() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}
	fasthttp.ServeFile(ctx, "./websocket.js")
}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			serveHome(ctx)
		case "/websocket.js":
			js(ctx)
		case "/ws":
			serveWs(ctx, hub)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	server := fasthttp.Server{
		Name:    "ChatExample",
		Handler: requestHandler,
	}

	fmt.Println("Starting server on", *addr)
	// Start the server
	//log.Fatal(server.ListenAndServeTLS(*addr, "./server.crt", "./server.key"))
	log.Fatal(server.ListenAndServe(*addr))
}

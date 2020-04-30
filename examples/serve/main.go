package main

import "github.com/whyiyhw/gws"

func main()  {
	server := new(gws.Server)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

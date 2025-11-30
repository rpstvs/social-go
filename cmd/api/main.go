package main

import "log"

func main() {

	const Addr = ":8080"

	config := NewConfig(Addr)

	app := NewApplication(config)

	mux := app.mount()

	log.Fatal(app.run(mux))
}

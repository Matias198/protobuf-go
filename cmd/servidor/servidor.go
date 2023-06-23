package main

import (
	"flag"
	"fmt"
	mensajero "mensajero/pkg"

	"google.golang.org/grpc"
)

func main() {

	// para argumento -p puerto
	punteroPuertoServidor := flag.String("p", "12345", "puerto del servidor")
	flag.Parse()
	fmt.Println(*punteroPuertoServidor)

	listen, port, err := mensajero.AbrirListener(*punteroPuertoServidor)
	fmt.Println("Escuchando en el puerto ", port)

	if err != nil {
		fmt.Println(err)
		return
	}

	servicioMensajero := mensajero.NuevoServidor()

	servidorReal := grpc.NewServer(
		grpc.UnaryInterceptor(servicioMensajero.Interceptor),
	)
	mensajero.RegisterMensajeroServer(servidorReal, &servicioMensajero)
	if err := servidorReal.Serve(listen); err != nil {
		fmt.Println("Falla: ", err)
	}
}

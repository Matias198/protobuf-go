package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	mensajero "mensajero/pkg"
)

const (
	DIRECCION_SERVIDOR_PREDETERMINADA       string = "localhost"
	PUERTO_SERVIDOR_PREDETERMINADO          string = "12345"
	USUARIO_PREDETERMINADO                  string = "jeffra"
	TEMPORIZADOR_EN_SEGUNDOS_PREDETERMINADO int    = 3
)

func main() {

	punteroUsuario := flag.String("u", "", "nombre de usuario usado por el cliente")
	punteroPuertoServidor := flag.String("p", "", "puerto a conectarse")
	punteroDireccionServidor := flag.String("d", "", "dirección del servidor")
	flag.Parse()

	iniciar(*punteroUsuario, *punteroPuertoServidor, *punteroDireccionServidor)
}

func iniciar(usuario string, puertoServidor string, direccionServidor string) {

	if usuario == "" {
		usuario = USUARIO_PREDETERMINADO
	}
	if direccionServidor == "" {
		direccionServidor = DIRECCION_SERVIDOR_PREDETERMINADA
	}
	if puertoServidor == "" {
		puertoServidor = PUERTO_SERVIDOR_PREDETERMINADO
	}

	direccion := fmt.Sprintf("%s:%s", direccionServidor, puertoServidor)
	conexion, cliente, ctx, err := mensajero.ConfigurarCliente(direccion, usuario, TEMPORIZADOR_EN_SEGUNDOS_PREDETERMINADO)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conexion.Close()

	fmt.Printf("Bienvenido %s. Pruebe cualquiera de los siguientes comandos\n", usuario)
	fmt.Println("\t obtener - ver los nuevos mensajes desde la última actualización")
	fmt.Println("\t listar - ver todos los usuarios conectados")
	fmt.Println("\t salir - Se desconecta")
	fmt.Println("\t <usuario> <mensaje...> - Envía <mensaje> al <usuario>")

	for {
		fmt.Printf("%s@ ", usuario)
		lector := bufio.NewReader(os.Stdin)
		linea, _ := lector.ReadString('\n')
		linea = strings.TrimSpace(linea)
		args := strings.SplitN(linea, " ", 2)

		respuesta, err := mensajero.Ejecutar(cliente, ctx, args...)
		if err != nil {
			fmt.Println(err)
			switch err.(type) {
			case *mensajero.ErrorDesconexion:
				return
			default:
			}
		}
		fmt.Println(respuesta)
	}
}

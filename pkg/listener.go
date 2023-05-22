package pkg

// REVISAR USO

import (
	"math/rand"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

// rango de puertos
const MENOR_PUERTO int = 32768
const MAYOR_PUERTO int = 61000

// Errno para dar soporte a los que usan Windows
const WIN_EADDRINUSE = syscall.Errno(10048)

// Escucha en un puerto aleatorio en el rango definido, 
// vuelve a intentarlo si el puerto ya está en uso
func AbrirListener(puertoSolicitado string) (net.Listener, string, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	puerto := strconv.Itoa(rand.Intn(MAYOR_PUERTO - MENOR_PUERTO) + MENOR_PUERTO)
	if puertoSolicitado != "" {
		puerto = puertoSolicitado
	}
	conexion, err := net.Listen("tcp", ":"+puerto)
	if err != nil {
		if direccionEnUso(err) {
			time.Sleep(100 * time.Millisecond)
			return AbrirListener("")
		} else {
			return nil, "", err
		}
	}
	return conexion, puerto, err
}

func direccionEnUso(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if osErr, ok := opErr.Err.(*os.SyscallError); ok {
			return osErr.Err == syscall.EADDRINUSE || osErr.Err == WIN_EADDRINUSE
		}
	}
	return false
}

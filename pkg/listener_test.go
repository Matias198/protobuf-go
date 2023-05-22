package pkg

import (
	"net"
	"testing"
)

func TestErrorDireccion(t *testing.T) {
	// Abre el listener (con suerte) en un puerto no utilizado
	_, err := net.Listen("tcp", ":12345")
	if err != nil {
		t.Errorf("Error desconocido, posiblemente 12345 se encuentra en uso: %s", err)
	}

	// Vuelve a abrir el listener en el puerto abierto anteriormente, 
	// esto debería causar un error EADDRINUSE
	_, err = net.Listen("tcp", ":12345")

	// Prueba que nuestro detector EADDRINUSE funciona
	if !direccionEnUso(err) {
		t.Errorf("Debería detectar correctamente el error EADDRINUSE!")
	}
}

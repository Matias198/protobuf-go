package mensajero

import (
	"fmt"
	"math/rand"
	mensajero "mensajero/pkg"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func enteroAleatorio(minimo int, maximo int) int {
	return minimo + rand.Intn(maximo-minimo)
}

// genera una cadena aleatoria de cierto largo
func stringAleatorio(largo int) string {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, largo)
	for i := 0; i < largo; i++ {
		bytes[i] = byte(enteroAleatorio(97, 122))
	}
	return string(bytes)
}

// Prueba que los tokens de autenticación se emiten y almacenan cuando un cliente contacta
// un servidor.
func TestAutenticacionUsuarioUnico(t *testing.T) {

	usuario := stringAleatorio(12)

	servicioMensajero := mensajero.NuevoServidor()
	servidorReal := grpc.NewServer(
		grpc.UnaryInterceptor(servicioMensajero.Interceptor),
	)

	if len(servicioMensajero.TablaAutenticacionUsuario) != 0 {
		t.Errorf("Se esperaba un elemento en TablaAutenticacionUsuario, encontrado %+v", servicioMensajero.TablaAutenticacionUsuario)
	}

	listen, puerto, _ := mensajero.AbrirListener("")
	direccion := fmt.Sprintf("localhost:%s", puerto)

	go func() {
		mensajero.RegisterMensajeroServer(servidorReal, &servicioMensajero)
		if err := servidorReal.Serve(listen); err != nil {
			t.Fatalf(err.Error())
		}
	}()

	defer func() {
		servidorReal.GracefulStop()
	}()

	conexion, _, ctx, err := mensajero.ConfigurarCliente(direccion, usuario, 3)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer conexion.Close()

	if len(servicioMensajero.TablaAutenticacionUsuario) != 1 {
		t.Errorf("Se esperaba un elemento en TablaAutenticacionUsuario, encontrado %+v", servicioMensajero.TablaAutenticacionUsuario)
	}

	for valor := range servicioMensajero.TablaAutenticacionUsuario {
		if valor != usuario {
			t.Errorf("El Usuario %s debe estar en la tabla de usuarios, pero la tabla de usuarios tenía %+v", usuario, servicioMensajero.TablaAutenticacionUsuario)
		}
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Errorf("Se esperaba que el contexto tuviera metadatos, se obtuvo %+v", ctx)
	}

	if _, recuperado := md["token"]; !recuperado {
		t.Errorf("Se esperaba que el contexto tuviera el token, se obtuvo %+v", md)
	}
}

// Probar que un solo cliente puede interactuar con el servidor
func TestUnSoloClienteInteractua(t *testing.T) {

	usuario := stringAleatorio(12)
	servicioMensajero := mensajero.NuevoServidor()
	servidorReal := grpc.NewServer(
		grpc.UnaryInterceptor(servicioMensajero.Interceptor),
	)

	listen, puerto, _ := mensajero.AbrirListener("")
	direccion := fmt.Sprintf("localhost:%s", puerto)

	go func() {
		mensajero.RegisterMensajeroServer(servidorReal, &servicioMensajero)
		if err := servidorReal.Serve(listen); err != nil {
			t.Fatalf(err.Error())
		}
	}()

	defer func() {
		servidorReal.GracefulStop()
	}()

	conexion, cliente, ctx, err := mensajero.ConfigurarCliente(direccion, usuario, 3)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer conexion.Close()

	respuesta, err := mensajero.Ejecutar(cliente, ctx, "listar")
	esperado := fmt.Sprintf("%s\n", usuario)
	if respuesta != esperado || err != nil {
		t.Errorf("Se esperaba %q en la llamada a `listar`, se obtuvo %q con error %+v", esperado, respuesta, err)
	}

	mensajero.Ejecutar(cliente, ctx, usuario, "hola")
	mensaje, err := mensajero.Ejecutar(cliente, ctx, "obtener")
	esperado = fmt.Sprintf("[%s]: hola\n", usuario)
	if mensaje != esperado || err != nil {
		t.Errorf("Se esperaba %q en la llamada a `obtener` después de un mensaje, se obtuvo %q con error %+v", esperado, mensaje, err)
	}

	mensajero.Ejecutar(cliente, ctx, usuario, "mensaje de varias partes 1")
	mensajero.Ejecutar(cliente, ctx, usuario, "mensaje de varias partes 2")

	mensajes, err := mensajero.Ejecutar(cliente, ctx, "obtener")
	esperado = fmt.Sprintf("[%s]: mensaje de varias partes 1\n[%s]: mensaje de varias partes 2\n", usuario, usuario)
	if mensajes != esperado || err != nil {
		t.Errorf("Se esperaba %q en la llamada a `obtener` después de varios mensajes, se obtuvo %q con error %+v", esperado, mensaje, err)
	}

	esperado = ""
	for i := 0; i < 2*mensajero.LARGO_LOTE; i++ {
		mensajero.Ejecutar(cliente, ctx, usuario, fmt.Sprintf("%d", i))
		if i < mensajero.LARGO_LOTE {
			esperado += fmt.Sprintf("[%s]: %d\n", usuario, i)
		}
	}
	mensajes, err = mensajero.Ejecutar(cliente, ctx, "obtener")
	if mensajes != esperado {
		t.Errorf("Se esperaba %s, se obtuvo %s al solicitar más mensajes que el largo del lote", esperado, mensajes)
	}
}

// Probar que varios clientes pueden interactuar con el servidor
func TestMultipleClients(t *testing.T) {

	usuario1 := stringAleatorio(12)
	time.Sleep(1 * time.Nanosecond)
	usuario2 := stringAleatorio(12)

	servicioMensajero := mensajero.NuevoServidor()
	servidorReal := grpc.NewServer(
		grpc.UnaryInterceptor(servicioMensajero.Interceptor),
	)

	listen, puerto, _ := mensajero.AbrirListener("")
	direccion := fmt.Sprintf("localhost:%s", puerto)

	go func() {
		mensajero.RegisterMensajeroServer(servidorReal, &servicioMensajero)
		if err := servidorReal.Serve(listen); err != nil {
			t.Fatalf(err.Error())
		}
	}()

	defer func() {
		servidorReal.GracefulStop()
	}()

	conexion1, cliente1, ctx1, err := mensajero.ConfigurarCliente(direccion, usuario1, 3)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer conexion1.Close()

	conexion2, cliente2, ctx2, err := mensajero.ConfigurarCliente(direccion, usuario2, 3)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer conexion2.Close()

	respuesta, err := mensajero.Ejecutar(cliente1, ctx1, "listar")
	// pueden implementar esto de manera independiente del orden
	esperadoA := fmt.Sprintf("%s, %s\n", usuario1, usuario2)
	esperadoB := fmt.Sprintf("%s, %s\n", usuario2, usuario1)
	if (respuesta != esperadoA && respuesta != esperadoB) || err != nil {
		t.Errorf("Se esperaba %q o %q en la llamada a `listar`, se obtuvo %q con error %+v", esperadoA, esperadoB, respuesta, err)
	}

	esperado := ""
	for i := 0; i < 2*mensajero.LARGO_LOTE; i++ {
		mensajero.Ejecutar(cliente2, ctx2, usuario1, fmt.Sprintf("%d", i))
		if i < mensajero.LARGO_LOTE {
			esperado += fmt.Sprintf("[%s]: %d\n", usuario2, i)
		}
	}

	mensajes, err := mensajero.Ejecutar(cliente1, ctx1, "obtener")
	if mensajes != esperado {
		t.Errorf("Se esperaba %s, se obtuvo %s al solicitar más mensajes que el largo del lote", esperado, mensajes)
	}

	println("Esperando 5 segundos para que los mensajes se entreguen a los usuarios")

}

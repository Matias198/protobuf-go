package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ErrorDesconexion struct {
	RazonesAdicionales string
}

func (e *ErrorDesconexion) Error() string {
	return fmt.Sprintf("El servidor se ha desconectado: %s", e.RazonesAdicionales)
}

/*
Regístrese como nuevo usuario con el servidor activo.
Obtenga el token de autenticación usando cliente.Conectar() y guárdelo en un objeto de `contexto`.
Se valida este objeto de contexto en el lado del servidor.

TODO: Implementar `Registrar`. Debe llamar a la RPC `Conectar` y usar el paquete `metadata`
apropiadamente para colocar el token de autenticación devuelto en un objeto context.Context.

Mire al interceptor del lado del servidor en `servidor_nucleo.go` para comprender cómo debería ser
la estructura del objeto context.Context.

Si se produce algún error, devuelva el mensaje de error que desee.
*/
func Registrar(cliente MensajeroClient, usuario string) (context.Context, error) {
	// Debe llamar a la RPC `Conectar` y usar el paquete `metadata`
	token, err := cliente.Conectar(context.Background(), &Registracion{
		UsuarioOrigen: usuario,
	})
	if err != nil {
		return nil, fmt.Errorf("no se puede registrar con el servidor: %s", err)
	}

	md := metadata.New(map[string]string{
		"nombreusuario": usuario,
		"token":         token.Token,
	})

	fmt.Printf("Token: %s\nUsuario: %s\n", md["token"], md["nombreusuario"])

	// Crear un nuevo contexto con el objeto metadata.MD
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return ctx, nil
}

// Una función auxiliar que devuelve una conexión de cliente activa con el servidor.
func ConfigurarCliente(direccion string, usuario string, temporizador int) (*grpc.ClientConn, MensajeroClient, context.Context, error) {

	// Establece una conexión con el servidor
	temporizadorEnSegundos := time.Duration(temporizador) * time.Second
	ctx, cancelar := context.WithTimeout(context.Background(), temporizadorEnSegundos)
	defer cancelar()
	conexion, err := grpc.DialContext(
		ctx,
		direccion,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return &grpc.ClientConn{}, nil, nil, fmt.Errorf("no se puede conectar con el servidor: %s", err)
	}

	cliente := NewMensajeroClient(conexion)

	// registra el cliente como un nuevo usuario
	ctx, err = Registrar(cliente, usuario)
	if err != nil {
		return &grpc.ClientConn{}, nil, nil, fmt.Errorf("no se puede registrar con el servidor: %s", err)
	}

	return conexion, cliente, ctx, nil
}

// Una función auxiliar que lleva a cabo las acciones indicadas por los argumentos.
// Los argumentos pueden ser un slice de cadena de uno o dos elementos.
// Si contiene dos elementos, el cliente envía un mensaje al servidor:
// el primer elemento se trata como el usuario al que se envía y
// el segundo elemento es el mensaje completo que se envía.
// Devuelve una cadena para mostrar al usuario los resultados de la operación.
func Ejecutar(cliente MensajeroClient, ctx context.Context, argumentos ...string) (string, error) {

	if len(argumentos) == 1 {
		switch argumentos[0] {
		case "obtener":

			mensajes, err := cliente.Obtener(ctx, &Vacio{})
			if err != nil {
				return "", err
			}

			todos := []string{}
			for _, mensaje := range mensajes.Mensajes {
				todos = append(todos, fmt.Sprintf("[%s]: %s", mensaje.Usuario, mensaje.Cuerpo))
			}

			return fmt.Sprintf("%s\n", strings.Join(todos, "\n")), nil

		case "listar":
			// TODO: ¡Implemente la llamada RPC del cliente para listar!
			// Esto debería mostrar una cadena separada por comas de todos los usuarios
			// devueltos por la RPC, que termina con un carácter de nueva línea "\n".
			// El orden de los usuarios impresos no importa.
			// Si se produce algún error, devuelva el mensaje de error que desee.
			usuarios, err := cliente.Listar(ctx, &Vacio{})
			if err != nil {
				return "", err
			}

			todos := []string{}
			for _, usuario := range usuarios.Usuarios {
				todos = append(todos, fmt.Sprintf("%s", usuario))
			}

			return fmt.Sprintf("%s\n", strings.Join(todos, ", ")), nil

		case "salir":
			correcto, err := cliente.Desconectar(ctx, &Vacio{})

			if err != nil || !correcto.Ok {
				return "", &ErrorDesconexion{RazonesAdicionales: err.Error()}
			}
			if correcto.Ok {
				// Retornar un mensaje de éxito y error nulo
				return "Desconectado\n", nil
			}
		}
	}

	if len(argumentos) == 2 {
		exitoso, err := cliente.Enviar(ctx, &MensajeApp{
			Usuario: argumentos[0],
			Cuerpo:  argumentos[1],
		})

		if err != nil {
			return "", fmt.Errorf("error al enviar: %s", err)
		}

		if exitoso.Ok {
			// Retrona un mensaje de éxito y error nulo
			return fmt.Sprintf("Mensaje enviado a %s\n", argumentos[0]), nil
		}

	}

	return "", nil

}

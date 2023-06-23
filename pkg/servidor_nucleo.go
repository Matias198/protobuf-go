package pkg

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const LARGO_LOTE = 50
const LARGO_BUZON = 1024

// Una clave para usar con context.Context
type clave string

const (
	nombreUsuario clave = "nombreUsuario"
	token         clave = "token"
)

// Una función hash para Conectar, úsela para generar nuevos tokens.
// No se usa en ningún otro lugar.
func hash(nombre string) (resultado string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(nombre)))
}

// La implementación del servidor
type Servidor struct {
	// Un UnimplementedMensajeroServer que se puede usar para registrar el servidor con gRPC
	// y que se puede usar para implementar métodos de servidor adicionales. No es necesario
	UnimplementedMensajeroServer

	// Un mapa de tokens de autenticación
	TablaAutenticacionUsuario map[string]string
	// Un mapa de los usuarios a los mensajes en su bandeja de entrada.
	// La bandeja de entrada está modelada como un canal de tamaño MAILBOX_SIZE.
	BandejasEntrada map[string](chan *MensajeApp)
}

func NuevoServidor() Servidor {
	return Servidor{
		TablaAutenticacionUsuario: make(map[string]string),
		BandejasEntrada:           make(map[string](chan *MensajeApp)),
	}
}

// Un interceptor del lado del servidor que asigna los tokens de autenticación en nuestro `contexto` a los nombres de usuario.
// Rechaza las llamadas si no tienen un token de autenticación válido. Nota: hemos hecho nuestro interceptor
// en este caso un método en nuestra estructura del Servidor para que pueda tener acceso a las variables privadas del Servidor
// - sin embargo, este no es un requisito estricto para los interceptores en general.
func (s Servidor) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (respuesta interface{}, err error) {

	// permite que las llamadas al punto final de Conectar pasen
	if info.FullMethod == "/mensajero.Mensajero/Conectar" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, errors.New("no se pudieron leer los metadatos de la solicitud")
	}

	// Imprime los parametros ok y mdlength para depurar que esta leyendo el interceptor
	fmt.Printf("Lectura de metadatos en la peticion: %v\n", ok)
	if ok {
		//Recorre la tabla de autenticacion viendo si el usuario con el token estan conectados
		for usuario, token := range s.TablaAutenticacionUsuario {
			if md["token"][0] == token {
				// Imprime si encuentra el usuario con su token
				fmt.Printf("El usuario %s se encuentra conectado y autenticado\n", usuario)
				fmt.Println(info.FullMethod)
				return handler(context.WithValue(ctx, nombreUsuario, usuario), req)
			}
		}
	}

	return nil, errors.New("no se pudo obtener el usuario del token de autenticación, si se proporcionó")
}

// Implementación de Conectar definido en el archivo `.proto`.
// Convierte el nombre de usuario proporcionado por `Registracion` en un objeto `TokenAutenticacion`.
// El token devuelto es único para el usuario; si el usuario ya inició sesión,
// la conexión debe ser rechazada. Esta función crea una entrada correspondiente
// en `s.TablaAutenticacionUsuario` y `s.BandejasEntrada`.
func (s Servidor) Conectar(_ context.Context, r *Registracion) (*TokenAutenticacion, error) {

	// Generar un token de autenticación usando la función hash
	token := hash(r.UsuarioOrigen)

	// Guardar el token de autenticación en la tabla del servidor
	s.TablaAutenticacionUsuario[r.UsuarioOrigen] = token

	// Crear una bandeja de entrada para el usuario
	s.BandejasEntrada[r.UsuarioOrigen] = make(chan *MensajeApp, LARGO_BUZON)

	// Imprime el token de auternticacion
	fmt.Printf("El usuario %s se conectó\n", r.UsuarioOrigen)

	// Imprime la tabla de autenticacion actual
	fmt.Println("Tabla de autenticacion actual:")
	for token, usuario := range s.TablaAutenticacionUsuario {
		fmt.Printf("\t%s: %s\n", token, usuario)
	}

	// Devolver el token de autenticación al cliente
	return &TokenAutenticacion{
		Token: token,
	}, nil

}

// Implementación de Enviar definido en el archivo `.proto`.
// Debe escribir el mensaje de chat en la bandeja de entrada privada de un usuario de
// destino en s.BandejasEntrada.
//
// El mensaje de chat debe tener su campo 'Usuario' reemplazado con el usuario remitente
// (cuando lo reciba inicialmente, tendrá el nombre del destinatario en su lugar).
//
// Sugerencia: ¿no está seguro de cómo obtener el "usuario remitente"?  Consulte algunos
// de los códigos de plantilla proporcionados en este archivo.
//
// TODO: Implementar `Enviar`. Si se produce algún error, devuelva el mensaje de error
// que desee.
func (s Servidor) Enviar(ctx context.Context, msg *MensajeApp) (*Correcto, error) {

	/*
		Teniendo en quenta que el cliente utilza el metodo Enviar de la siguiente manera:
		exitoso, err := cliente.Enviar(ctx, &MensajeApp{
			Usuario: argumentos[0],
			Cuerpo:  argumentos[1],
		})
	*/

	// canal de destino
	destino := msg.Usuario
	canal, ok := s.BandejasEntrada[destino]
	if !ok {
		return nil, errors.New("el usuario destino no se encuentra conectado")
	}

	// Obtengo los metadatos de la solicitud
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, errors.New("no se pudieron leer los metadatos de la solicitud")
	}

	// Obtengo el usuario desde los metadatos
	usuario := md["nombreusuario"][0]

	msg.Usuario = usuario
	// envio del mensaje
	canal <- msg
	return &Correcto{Ok: true}, nil
}

// Implementación de Obtener definido en el archivo `.proto`.
// Debe consumir y devolver un número máximo de mensajes de acuerdo a LARGO_LOTE
// del canal de bandeja de entrada para el usuario actual.
//
// Sugerencia: use sentencias `select` en un bucle `for` adecuado para consumir del
// canal mientras haya mensajes restantes.
//
// TODO: Implementar Obtener. Si se produce algún error, devuelva el mensaje de error
// que desee.
func (s Servidor) Obtener(ctx context.Context, _ *Vacio) (*MensajesApp, error) {

	// Obtengo los metadatos de la solicitud
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, errors.New("no se pudieron leer los metadatos de la solicitud")
	}

	// Obtengo el usuario desde los metadatos
	usuario := md["nombreusuario"][0]

	// canal de destino
	canal, ok := s.BandejasEntrada[usuario]
	if !ok {
		return nil, errors.New("el usuario destino no se encuentra conectado")
	}

	// mensajes
	mensajes := &MensajesApp{
		Mensajes: []*MensajeApp{},
	}

	// obtencion de mensajes
	for i := 0; i < LARGO_LOTE; i++ {
		select {
		case msg := <-canal:
			mensajes.Mensajes = append(mensajes.Mensajes, msg)
		default:
			break
		}
	}

	return mensajes, nil
}

// Implementación de Listar definido en el archivo `.proto`.
// Debe devolver el listado de usuarios al momento de la llamada.
func (s *Servidor) Listar(ctx context.Context, vacio *Vacio) (*ListaUsuarios, error) {

	// Obtener el listado de usuarios registrados en el servidor
	usuarios := s.ObtenerUsuariosRegistrados()

	// Imprimir el listado de usuarios
	fmt.Println("Listado de usuarios registrados:")
	for _, usuario := range usuarios {
		fmt.Printf("\t%s\n", usuario)
	}

	// Crear la respuesta con el listado de usuarios
	respuesta := &ListaUsuarios{
		Usuarios: usuarios,
	}

	// Si hay un error, devolverlo
	if respuesta == nil {
		return nil, errors.New("No se pudo obtener el listado de usuarios")
	}

	return respuesta, nil
}

// Método para obtener el listado de usuarios registrados en el servidor
func (s *Servidor) ObtenerUsuariosRegistrados() []string {

	// Obtener el listado de usuarios registrados en el servidor
	usuarios := make([]string, 0, len(s.TablaAutenticacionUsuario))

	// Iterar sobre la tabla de autenticación para obtener los usuarios
	for usuario := range s.TablaAutenticacionUsuario {
		usuarios = append(usuarios, usuario)
	}

	// Devolver el listado de usuarios
	return usuarios
}

// Implementación de Desconectar definido en el archivo `.proto`.
// Debe destruir la bandeja de entrada correspondiente y la entrada en `s.TablaAutenticacionUsuario`.
func (s Servidor) Desconectar(ctx context.Context, _ *Vacio) (*Correcto, error) {

	// Obtengo los metadatos de la solicitud
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, errors.New("no se pudieron leer los metadatos de la solicitud")
	}

	// Obtengo el usuario desde los metadatos
	usuario := md["nombreusuario"][0]

	delete(s.TablaAutenticacionUsuario, usuario)
	delete(s.BandejasEntrada, usuario)

	return &Correcto{Ok: true}, nil
}

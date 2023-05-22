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


// Una función hash para Conectar, úsela para generar nuevos tokens.
// No se usa en ningún otro lugar.
func hash(nombre string) (resultado string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(nombre)))
}

// La implementación del servidor
type Servidor struct {
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
	fmt.Println(info.FullMethod)
	// permite que las llamadas al punto final de Conectar pasen
	if info.FullMethod == "/mensajero.Mensajero/Conectar" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no se pudieron leer los metadatos de la solicitud")
	}

	// si el token está presente en los metadatos
	if valores, ok := md["token"]; ok {
		if len(valores) == 1 {
			// si el usuario se encuentra presente en s.TablaAutenticacionUsuario
			if usuario, ok := s.TablaAutenticacionUsuario[valores[0]]; ok {
				return handler(context.WithValue(context.Background(), "nombreUsuario", usuario), req)
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

	token := hash(r.UsuarioOrigen)

	if _, ok := s.TablaAutenticacionUsuario[token]; !ok {
		s.TablaAutenticacionUsuario[token] = r.UsuarioOrigen
		s.BandejasEntrada[r.UsuarioOrigen] = make(chan *MensajeApp, LARGO_BUZON)

		return &TokenAutenticacion{
			Token: token,
		}, nil
	}

	return nil, errors.New("El usuario se encuentra conectado")

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
	return nil, nil
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
	return nil, nil
}

// Implementación de Listar definido en el archivo `.proto`.
// Debe devolver el listado de usuarios al momento de la llamada.
func (s Servidor) Listar(ctx context.Context, _ *Vacio) (*ListaUsuarios, error) {

	u := &ListaUsuarios{
		Usuarios: []string{},
	}

	for _, usuario := range s.TablaAutenticacionUsuario {
		u.Usuarios = append(u.Usuarios, usuario)
	}

	return u, nil

}

// Implementación de Desconectar definido en el archivo `.proto`.
// Debe destruir la bandeja de entrada correspondiente y la entrada en `s.TablaAutenticacionUsuario`.
func (s Servidor) Desconectar(ctx context.Context, _ *Vacio) (*Correcto, error) {
	usuario := fmt.Sprintf("%v", ctx.Value("nombreUsuario"))
	close(s.BandejasEntrada[usuario]) // se asegura de que no se puedan enviar más escrituras en este canal
	delete(s.BandejasEntrada, usuario)

	for token, u := range s.TablaAutenticacionUsuario {
		if u == usuario {
			delete(s.TablaAutenticacionUsuario, token)
		}
	}

	return &Correcto{Ok: true}, nil
}

# protobuf-go

# Laboratorio 3: Mensajero

### Objetivo

Cree un servidor y un cliente gRPC que se comuniquen correctamente entre sí.

### Metas

- Escribir un archivo proto que defina un servicio RPC
- Generar archivos go desde un archivo proto
- Enviar solicitudes de gRPC del cliente al servidor

<h2>Parte 1: Proto</h2>
<ol>
<li>Instalar protoc en su máquina local</li>
<li>Cree su archivo protobuf. Asegúrese de nombrar el paquete igual que sus otros archivos</li>
<li>Ejecute <code>make build</code> en el directorio del proyecto para generar los archivos go. </li>
</ol>

<h2>Parte 2: Servidor</h2>
Implementar los TODOs

<h2>Parte 3: Cliente</h2>
Implementar los TODOs

## Construir binarios

- Vaya a `cmd/cliente` o `cmd/servidor` y ejecute `go build .`. Esto generará archivos binarios de cliente o servidor respectivamente.

## Pruebas

- `go test ./...` debería ejecutar todas las pruebas.

## Compilar el .proto

- protoc --go_out=. --go-grpc_out=. ./pkg/*.proto

> Nota: Sin plugins 
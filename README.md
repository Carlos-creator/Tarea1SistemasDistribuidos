# Tarea 1: Intrusos en Rishi — INF-343 Sistemas Distribuidos

## Integrantes

| Nombre | Apellido | Rol |
|--------|----------|-----|
| Carlos | Ramírez | 202192826-k |
| [Nombre 2] | [Apellido 2] | [ROL-2] |
| [Nombre 3] | [Apellido 3] | [ROL-3] |

---

## Descripción

Sistema de inventario de armamento con dos implementaciones:

- **Parte 1**: API RESTful con almacenamiento en memoria volátil (los datos se pierden al reiniciar).
- **Parte 2**: API RESTful con almacenamiento persistente en SQLite, distribuida en 3 máquinas virtuales.

---

## Asignación de máquinas virtuales (Grupo 4)

| Rol | IP | Descripción |
|-----|----|-------------|
| MV1 | `10.10.28.20` | Cliente CLI |
| MV2 | `10.10.28.21` | Servidor API RESTful (Gin) |
| MV3 | `10.10.28.22` | Servicio de base de datos (SQLite) |

---

## Estructura del proyecto

```
tarea1/
├── go.mod
├── go.sum
├── parte1/
│   └── server/
│       └── main.go        # Servidor con memoria volátil (MV2)
├── parte2/
│   ├── server/
│   │   └── main.go        # Servidor proxy a BD (MV2)
│   └── dbservice/
│       └── main.go        # Servicio SQLite (MV3)
├── client/
│   └── main.go            # Cliente CLI (MV1)
└── testing/
    └── main.go            # Herramienta de benchmark (MV1)
```

---

## Instrucciones de compilación

Desde la raíz del proyecto, compilar los binarios para Linux (Ubuntu):

```bash
GOOS=linux GOARCH=amd64 go build -o bin/server_p1  ./parte1/server/
GOOS=linux GOARCH=amd64 go build -o bin/server_p2  ./parte2/server/
GOOS=linux GOARCH=amd64 go build -o bin/dbservice  ./parte2/dbservice/
GOOS=linux GOARCH=amd64 go build -o bin/client     ./client/
GOOS=linux GOARCH=amd64 go build -o bin/testing    ./testing/
```

Subir a las VMs con `scp`:

```bash
# MV3: servicio de base de datos
scp bin/dbservice ubuntu@10.10.28.22:~

# MV2: servidores
scp bin/server_p1 bin/server_p2 ubuntu@10.10.28.21:~

# MV1: cliente y herramienta de testing
scp bin/client bin/testing ubuntu@10.10.28.20:~
```

---

## Instrucciones de ejecución

### Parte 1 — Memoria volátil

```bash
# MV2: iniciar servidor
./server_p1 -port 8080

# MV1: iniciar cliente
./client -server http://10.10.28.21:8080
```

### Parte 2 — SQLite persistente

```bash
# MV3: iniciar servicio de base de datos
./dbservice -port 8081

# MV2: iniciar servidor (apunta a MV3)
./server_p2 -port 8080 -db http://10.10.28.22:8081

# MV1: iniciar cliente (misma interfaz)
./client -server http://10.10.28.21:8080
```

### Herramienta de benchmark

```bash
# MV1: testear Parte 1 (volatil)
./testing -server http://10.10.28.21:8080 -n 100 -output resultados_volatil.txt

# MV1: testear Parte 2 (persistente, con server_p2 corriendo en MV2)
./testing -server http://10.10.28.21:8080 -n 100 -output resultados_persistente.txt
```

**Flags disponibles:**

| Herramienta | Flag | Default | Descripción |
|-------------|------|---------|-------------|
| `server_p1` / `server_p2` | `-port` | `8080` | Puerto del servidor |
| `server_p2` | `-db` | `http://10.10.28.22:8081` | URL del servicio de BD |
| `dbservice` | `-port` | `8081` | Puerto del servicio de BD |
| `dbservice` | `-db-path` | `weapons.db` | Ruta del archivo SQLite |
| `client` | `-server` | `http://localhost:8080` | URL del servidor |
| `testing` | `-server` | `http://localhost:8080` | URL del servidor a testear |
| `testing` | `-n` | `100` | Peticiones por endpoint |
| `testing` | `-output` | `resultados.txt` | Archivo de salida |

---

## Endpoints de la API

| Método | Ruta | Body | Descripción |
|--------|------|------|-------------|
| `GET` | `/weapons` | — | Retorna el inventario completo |
| `POST` | `/weapons` | `{"weapon_name": "...", "stock": N}` | Registra un nuevo armamento |
| `PATCH` | `/weapons/:weapon_name` | `{"quantity": N}` | Retira N unidades del armamento |

**Códigos de respuesta:**
- `200 OK` — operación exitosa
- `201 Created` — armamento registrado
- `400 Bad Request` — datos inválidos o stock insuficiente
- `404 Not Found` — armamento no encontrado
- `409 Conflict` — el nombre del armamento ya existe

---

## Análisis de resultados de benchmark

> Completar con los valores reales obtenidos al ejecutar la herramienta de testing en las VMs.

### Tiempos de respuesta promedio

| Endpoint | Volátil (ms) | Persistente (ms) |
|----------|-------------|-----------------|
| `POST /weapons` | — | — |
| `GET /weapons` | — | — |
| `PATCH /weapons/:weapon_name` | — | — |

### Explicación de la diferencia

La versión con **memoria volátil** es más rápida porque todas las operaciones se realizan directamente sobre estructuras de datos en RAM (un `map` de Go protegido con mutex), sin ningún acceso a disco ni comunicación adicional por red.

La versión **persistente** presenta mayor latencia por dos razones acumuladas:

1. **I/O de disco**: SQLite escribe y lee datos desde el sistema de archivos de MV3. Las operaciones de escritura (`POST`, `PATCH`) requieren sincronización con disco para garantizar durabilidad.
2. **Latencia de red**: El servidor en MV2 debe realizar una petición HTTP adicional a MV3 por cada operación del cliente. Esto añade el tiempo de ida y vuelta entre ambas VMs (RTT de red).

### Decisión: ¿cuál solución es más conveniente?

Para el contexto planteado (inventario de armamento en zona de combate), la solución **persistente (Parte 2)** es claramente la más conveniente, pese a su mayor latencia, por las siguientes razones:

- **Durabilidad**: El evento del droide DRK-1 demuestra que cortes de energía son una amenaza real. Con memoria volátil, cualquier fallo implica pérdida total del inventario. La base de datos en MV3 sobrevive reinicios del servidor.
- **Disponibilidad del dato**: Almacenar los datos en una máquina separada (MV3) con menos riesgo de ataque, como propone Fives, protege el inventario ante compromisos físicos de MV2.
- **La diferencia de latencia es aceptable**: En este dominio, unos milisegundos adicionales por petición no son críticos frente a la pérdida total de información operacional en combate.

La memoria volátil solo sería preferible si la velocidad de respuesta fuera estrictamente crítica y existieran mecanismos alternativos de recuperación ante fallos.

---

## Consideraciones especiales

- El cliente es idéntico para ambas partes. El cambio de API es transparente: solo varía la URL del servidor con el flag `-server`.
- La herramienta de benchmark limita las peticiones PATCH a un máximo de 999 para no exceder el stock inicial de 1000 unidades.
- `modernc.org/sqlite` se usa en lugar de `github.com/mattn/go-sqlite3` para evitar la dependencia de CGO, facilitando la compilación cruzada para Linux desde Windows.
- Si se utiliza asistencia por IA: se usó Claude Code para la generación inicial de los archivos `parte1/server/main.go`, `parte2/server/main.go`, `parte2/dbservice/main.go`, `client/main.go` y `testing/main.go`. Se revisó y validó el código generado. Los comentarios automáticos de la herramienta fueron eliminados y reemplazados por comentarios propios del grupo.

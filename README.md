# Laboratorio 1: GitHub, Docker y API REST

## Integrantes
- Juan Barrero (A): Juan Barrero
- Sebastian Portes (B): Juan Diego Garzon

## Descripción
API REST en Go con PostgreSQL para gestionar notas de un equipo: Juan Barrero implementó los endpoints de consulta y Sebastian Portes implementó las operaciones de escritura.

## Tecnologías
- Go 1.25 o posterior; net/http y database/sql.
- Controlador github.com/lib/pq v1.10.9.
- PostgreSQL 17 en Docker.
- Docker y Docker Compose.
- Git y GitHub.

## Requisitos
- Docker Desktop funcionando con contenedores Linux.
- Git y curl.exe.
- PowerShell en Windows.
- Go 1.25 o posterior para comprobaciones locales.
- Puerto 8000 disponible.

## Ejecución desde una copia nueva
En PowerShell:

```powershell
git clone https://github.com/USUARIO_A/lab1-barrero-portes.git
Set-Location lab1-barrero-portes
Copy-Item .env.example .env
```

Editar .env antes de arrancar:
- Cambiar DB_PASSWORD por una contraseña local propia.
- Cambiar APP_ENV para identificar el entorno.
- Conservar DB_HOST=db y DB_PORT=5432.
- No subir .env a Git.

```powershell
docker compose config --quiet
docker compose up -d --build
docker compose ps
curl.exe -i http://localhost:8000/health
```

## Arquitectura
Cliente HTTP en Windows → API Go, servicio api → PostgreSQL, servicio db.
La API lee APP_ENV, APP_PORT, DB_HOST, DB_PORT, DB_NAME, DB_USER y DB_PASSWORD.
El puerto publicado de la API se limita a 127.0.0.1.
PostgreSQL se usa dentro de la red de Compose, sin publicar 5432.

## Base de datos
db/init.sql crea notes al inicializar un volumen vacío.
Campos: id, title, content, author y created_at.
id y created_at los genera PostgreSQL.
El volumen postgres_data se monta en /var/lib/postgresql/data.
El script inicial no vuelve a ejecutarse sobre una base ya inicializada.

## Endpoints
| Método | Ruta | Resultado |
|---|---|---|
| GET | /health | 200, salud y entorno |
| GET | /notes | 200, lista JSON, incluso vacía |
| GET | /notes/{id} | 200 o 404 |
| POST | /notes | 201 o 400 |
| PUT | /notes/{id} | 200, 400 o 404 |
| DELETE | /notes/{id} | 200 o 404 |

Todos los resultados y errores de la API usan JSON.
Los ids inválidos devuelven 400.
Los errores internos devuelven 500 con un mensaje genérico.
title, content y author son obligatorios y no admiten valores vacíos.
title admite hasta 200 caracteres y author hasta 100.
PUT conserva created_at; no crea notas inexistentes.

## Prueba rápida
```powershell
curl.exe -i -X POST http://localhost:8000/notes -H "Content-Type: application/json" --data-binary "@requests/create.json"
curl.exe -i http://localhost:8000/notes
```

Copiar el id real devuelto por POST en la variable siguiente:

```powershell
# Usar 1 solamente si el servidor devolvió id 1.
$id = 1
curl.exe -i "http://localhost:8000/notes/$id"
curl.exe -i -X PUT "http://localhost:8000/notes/$id" -H "Content-Type: application/json" --data-binary "@requests/update.json"
curl.exe -i -X POST http://localhost:8000/notes -H "Content-Type: application/json" --data-binary "@requests/invalid.json"
curl.exe -i -X DELETE "http://localhost:8000/notes/$id"
```

## Persistencia
Crear una nota con requests/persistence.json, anotar su id y consultar.

```powershell
docker compose down
docker compose up -d
curl.exe -i http://localhost:8000/notes
```

La nota debe conservarse. No usar down -v: elimina los datos del volumen.

## Comprobaciones locales
```powershell
go test ./...
go vet ./...
```

Si go test informa [no test files], solo se comprobó compilación:
las pruebas funcionales se ejecutan con curl y se registran en el informe.

## Trabajo colaborativo
- main: versión integrada por Pull Request.
- develop: integración de ramas de trabajo.
- A: GET /health, GET /notes y GET /notes/{id}.
- B: POST /notes, PUT /notes/{id} y DELETE /notes/{id}.
- Revisiones cruzadas con comentarios y respuestas.
- Conflicto en README resuelto por B conservando ambos aportes.

## Seguridad y alcance
Proyecto didáctico local sin autenticación.
No contiene credenciales reales versionadas.
La conexión local a PostgreSQL usa sslmode=disable.
No se debe publicar esta configuración como servicio de producción.
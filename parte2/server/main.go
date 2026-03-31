// Servidor Parte 2: API RESTful que actúa como intermediario entre el cliente
// y el servicio de base de datos SQLite que corre en MV3.
// Expone los mismos endpoints que el servidor de Parte 1, pero delega
// toda la lógica de datos al servicio de BD en MV3.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// dbURL es la URL base del servicio de BD en MV3.
var dbURL string

// proxyRequest reenvía la petición HTTP al servicio de BD y retorna su respuesta al cliente.
// Copia método, path, body y status code de forma transparente.
func proxyRequest(c *gin.Context, method, path string, body []byte) {
	url := dbURL + path

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creando request al servicio de BD"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no se pudo conectar al servicio de BD"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error leyendo respuesta del servicio de BD"})
		return
	}

	c.Data(resp.StatusCode, "application/json", respBody)
}

// getWeapons delega la consulta del inventario al servicio de BD.
func getWeapons(c *gin.Context) {
	proxyRequest(c, http.MethodGet, "/weapons", nil)
}

// postWeapon delega el registro de armamento al servicio de BD.
func postWeapon(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error leyendo body"})
		return
	}
	proxyRequest(c, http.MethodPost, "/weapons", body)
}

// patchWeapon delega el retiro de armamento al servicio de BD.
func patchWeapon(c *gin.Context) {
	name := c.Param("weapon_name")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error leyendo body"})
		return
	}
	proxyRequest(c, http.MethodPatch, "/weapons/"+name, body)
}

func main() {
	port := flag.Int("port", 8080, "puerto en el que escucha el servidor")
	db := flag.String("db", "http://10.10.28.22:8081", "URL del servicio de base de datos (MV3)")
	flag.Parse()

	dbURL = *db

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/weapons", getWeapons)
	r.POST("/weapons", postWeapon)
	r.PATCH("/weapons/:weapon_name", patchWeapon)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("[Parte 2] Servidor iniciado en %s -> BD en %s\n", addr, dbURL)
	if err := r.Run(addr); err != nil {
		fmt.Printf("Error al iniciar servidor: %v\n", err)
	}
}

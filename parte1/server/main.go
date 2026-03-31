// Servidor Parte 1: API RESTful con almacenamiento en memoria volátil.
// Implementa los endpoints GET /weapons, POST /weapons y PATCH /weapons/:weapon_name
// usando Gin Web Framework. Los datos se pierden al reiniciar el proceso.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Weapon representa un arma en el inventario.
type Weapon struct {
	ID         int    `json:"id"`
	WeaponName string `json:"weapon_name"`
	Stock      int    `json:"stock"`
}

// postBody es el cuerpo esperado para crear un arma.
type postBody struct {
	WeaponName string `json:"weapon_name" binding:"required"`
	Stock      int    `json:"stock"       binding:"required,min=1"`
}

// patchBody es el cuerpo esperado para retirar armamento.
type patchBody struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// store mantiene el inventario en memoria.
var (
	weapons = make(map[string]*Weapon)
	mu      sync.RWMutex
	nextID  = 1
)

// getWeapons devuelve todos los armamentos del inventario.
func getWeapons(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()

	list := make([]*Weapon, 0, len(weapons))
	for _, w := range weapons {
		list = append(list, w)
	}
	c.JSON(http.StatusOK, list)
}

// postWeapon registra un nuevo armamento en el inventario.
// Retorna 409 si el nombre ya existe.
func postWeapon(c *gin.Context) {
	var body postBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := weapons[body.WeaponName]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "el arma ya existe"})
		return
	}

	w := &Weapon{
		ID:         nextID,
		WeaponName: body.WeaponName,
		Stock:      body.Stock,
	}
	nextID++
	weapons[body.WeaponName] = w
	c.JSON(http.StatusCreated, w)
}

// patchWeapon descuenta unidades de un arma existente.
// Retorna 404 si el arma no existe y 400 si el stock es insuficiente.
func patchWeapon(c *gin.Context) {
	name := c.Param("weapon_name")

	var body patchBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	w, exists := weapons[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "arma no encontrada"})
		return
	}

	if body.Quantity > w.Stock {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stock insuficiente"})
		return
	}

	w.Stock -= body.Quantity
	c.JSON(http.StatusOK, w)
}

func main() {
	port := flag.Int("port", 8080, "puerto en el que escucha el servidor")
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/weapons", getWeapons)
	r.POST("/weapons", postWeapon)
	r.PATCH("/weapons/:weapon_name", patchWeapon)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("[Parte 1] Servidor iniciado en %s (memoria volatil)\n", addr)
	if err := r.Run(addr); err != nil {
		fmt.Printf("Error al iniciar servidor: %v\n", err)
	}
}

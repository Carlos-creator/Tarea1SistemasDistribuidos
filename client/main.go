package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Estructura que coincide con la entidad "Arma" definida [cite: 49-52]
type Weapon struct {
	ID         int    `json:"id"`
	WeaponName string `json:"weapon_name"`
	Stock      int    `json:"stock"`
}

// URL base del servidor (MV2). CAMBIAR
const baseURL = "http://10.10.28.21:8080/weapons"

func main() {
	for {
		fmt.Println("\nGESTIÓN DE ARMAMENTO")
		fmt.Println("1. Ver Inventario")
		fmt.Println("2. Añadir armas")
		fmt.Println("3. Retirar arma(s)")
		fmt.Println("4. Salir")
		fmt.Print("Selecciona una opción: ")

		var opcion int
		fmt.Scanln(&opcion)

		switch opcion {
		case 1:
			verInventario()
		case 2:
			añadirArmas()
		case 3:
			retirarArmas()
		case 4:
			fmt.Println("Saliendo del sistema...")
			os.Exit(0)
		default:
			fmt.Println("Opción no válida.")
		}
	}
}

// 1. Ver Inventario [cite: 46, 61]
func verInventario() {
	resp, err := http.Get(baseURL)
	if err != nil {
		fmt.Println("Error de conexión:", err)
		return
	}
	defer resp.Body.Close()

	var weapons []Weapon
	json.NewDecoder(resp.Body).Decode(&weapons)

	fmt.Println("\nNOMBRE\t\tCANTIDAD")
	fmt.Println("=======\t\t=======")
	for _, w := range weapons {
		fmt.Printf("%s\t\t%d\n", w.WeaponName, w.Stock)
	}
}

// 2. Añadir armas [cite: 44, 63]
func añadirArmas() {
	var name string
	var cantidad int

	fmt.Print("Nombre del arma: ")
	fmt.Scanln(&name)
	fmt.Print("Cantidad: ")
	fmt.Scanln(&cantidad)

	payload, _ := json.Marshal(map[string]interface{}{
		"weapon_name": name,
		"stock":       cantidad,
	})

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		fmt.Println("Error al añadir armamento.")
		return
	}
	fmt.Println("Armamento añadido con éxito.")
}

// 3. Retirar arma(s) [cite: 45, 62]
func retirarArmas() {
	var name string
	var cantidad int

	fmt.Print("¿Qué arma quieres descontar?: ")
	fmt.Scanln(&name)
	fmt.Print("Cantidad a retirar: ")
	fmt.Scanln(&cantidad)

	payload, _ := json.Marshal(map[string]int{"quantity": cantidad})
	
	// PATCH requiere una sintaxis especial en Go http.NewRequest
	client := &http.Client{}
	url := fmt.Sprintf("%s/%s", baseURL, name)
	req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Manejo de errores simplificado: stock insuficiente o no encontrada
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		return
	}
	fmt.Println("Retiro aplicado.")
}
/*
Copyright © 2025 NAME HERE axel.burling@gmail.com
*/
package main

import (
	"github.com/axelburling/dfs/cmd"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load(".env")
}

func main() {
	cmd.Execute()
}

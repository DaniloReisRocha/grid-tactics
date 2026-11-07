package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DaniloReisRocha/grid-tactics/backend/server/handler"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Destination path not found")
	}
	frontendPath := os.Args[1]

	zodStr, err := makeZodClient()
	if err != nil {
		log.Fatalf("Failed to generate zodios client: %v", err)
	}

	if str, e := formatTypeScriptCode(zodStr); e == nil {
		zodStr = str
	} else {
		log.Fatalf("Failed to format zodios: %v", e)
	}

	err = WriteStringToFile(zodStr, frontendPath)
	if err != nil {
		log.Fatalf("Fail WriteStringToFile: %v", err)
	}

	fmt.Printf("✅ Api client created in: %s", frontendPath)
}

func makeZodClient() (string, error) {
	handler := handler.HTTPServer{}

	var schemas strings.Builder
	var endpoints strings.Builder

	endpoints.WriteString("const endpoints = makeApi([\n")
	for _, ep := range handler.Routes(nil) {
		zodStr, err := ep.MakeZod()
		if err != nil {
			return "", err
		}
		schemas.WriteString(zodStr.Schemas + "\n")
		endpoints.WriteString(zodStr.Endpoint)
	}
	endpoints.WriteString("]);")

	var sb strings.Builder
	sb.WriteString("import { makeApi, Zodios, type ZodiosOptions } from \"@zodios/core\";\n")
	sb.WriteString("import { z } from \"zod\";\n\n")
	sb.WriteString(schemas.String())
	sb.WriteString("\n\n")
	sb.WriteString(endpoints.String())
	sb.WriteString("\n\n")
	sb.WriteString("export function createApiClient(baseUrl: string, options?: ZodiosOptions) {\n")
	sb.WriteString("\treturn new Zodios(baseUrl, endpoints, options);\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

func WriteStringToFile(content, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "create directory")
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return errors.Wrap(err, "write content")
	}

	return nil
}

func formatTypeScriptCode(code string) (string, error) {
	cmd := exec.Command("npx", "prettier", "--parser", "typescript")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	_, err = stdin.Write([]byte(code))
	if err != nil {
		return "", err
	}
	stdin.Close()

	formatted, err := io.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return string(formatted), nil
}

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	
	apiKey := "AIzaSyCpzTBSG1Oo_GtHSC69kwSoknCPX4xUHs4" 

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("Запрашиваем список моделей у Google...")
	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Ошибка при получении списка: %v", err)
		}
		if m.Name != "" {
			fmt.Printf("Найдена модель: %s\n", m.Name)
		}
	}
	fmt.Println("Готово!")
}
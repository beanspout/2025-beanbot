package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/NZ26RQ_gme/lsie-beanbot/internal/knowledge"
	"github.com/NZ26RQ_gme/lsie-beanbot/internal/ollama"
	"github.com/NZ26RQ_gme/lsie-beanbot/internal/ui"
)

func main() {
	// Initialize Fyne application
	myApp := app.NewWithID("com.example.beanbot")
	myWindow := myApp.NewWindow("BeanBot - LSIE Support")
	myWindow.Resize(fyne.NewSize(450, 700)) // Optimized chat window size

	// Initialize knowledge database
	kb, err := knowledge.NewKnowledgeDatabase()
	if err != nil {
		log.Fatal("Failed to initialize knowledge database:", err)
	}

	// Initialize Ollama client (model can be changed via footer click)
	ollamaClient := ollama.NewClient("http://localhost:11434", "gemma3:1b")

	// Initialize BeanBot UI
	bot := ui.NewBeanBot(myApp, myWindow, kb, ollamaClient)

	// Setup and display UI
	bot.SetupUI()
	myWindow.ShowAndRun()
}

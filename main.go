package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/nkcyber/ai-hacking-lab/components"
	"github.com/nkcyber/ai-hacking-lab/services"
	slogchi "github.com/samber/slog-chi"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	// configurable flags
	promptPath := flag.String("promptPath", "./example-prompts.json", "the filepath to load prompts from")
	modelName := flag.String("modelName", "tinyllama", "the name of the LLM in the Ollama library")
	modelTemperature := flag.Float64("modelTemperature", 0.1, "the 'temperature' of the LLM")
	maxTokens := flag.Int("maxTokens", 100, "the maximum number of tokens in a response.")
	address := flag.String("address", ":3000", "the address to host the server on")
	redisAddress := flag.String("redisAddress", "localhost:6379", "the address to connect to redis on")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// create chatbot
	chat, err := services.NewChat(*modelName, *modelTemperature, *maxTokens, *promptPath, *redisAddress, logger)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	err = chat.ClearAllMessages()
	if err != nil {
		logger.Error(err.Error())
	}
	// create router
	router := chi.NewRouter()
	router.Use(middleware.Timeout(120 * time.Second))
	router.Use(slogchi.New(logger))
	// TODO: shove all of this in a handlers package
	router.Route("/chat", func(r chi.Router) {
		r.Use(httprate.Limit(
			3,             // requests
			3*time.Second, // per duration
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		))
		workDir, _ := os.Getwd()
		filesDir := http.Dir(filepath.Join(workDir, "assets"))
		FileServer(r, "/assets", filesDir)
		r.Post("/{chatId}", func(w http.ResponseWriter, r *http.Request) {
			// ensure chat id exists
			chatId := services.ChatIdType(chi.URLParam(r, "chatId"))
			if len(chatId) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if err := chat.AssertChatIdExists(chatId); err != nil {
				logger.Warn(fmt.Sprintf("could not find chatId '%s'", chatId))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// get question from body
			err = r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			userInput := r.FormValue("message")
			if len(userInput) < 5 || len(userInput) > 200 {
				components.Response("", "Your message must be between 5 and 200 characters!").Render(r.Context(), w)
				return
			}
			logger.Info("RESPONDING to " + string(chatId))
			chat.AddMessage(chatId, schema.HumanChatMessage{Content: userInput})
			response, err := chat.Respond(chatId)
			if err != nil {
				logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			components.Response(userInput, response.GetContent()).Render(r.Context(), w)
		})
		r.Get("/{promptName}", func(w http.ResponseWriter, r *http.Request) {
			promptName := chi.URLParam(r, "promptName")
			if !chat.PromptExists(promptName) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			components.Index(promptName).Render(r.Context(), w)
		})
		r.Post("/start/{promptName}", func(w http.ResponseWriter, r *http.Request) {
			promptName := chi.URLParam(r, "promptName")
			if !chat.PromptExists(promptName) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			id, err := chat.StartChat(promptName)
			if err != nil {
				logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			components.StartChat(string(id)).Render(r.Context(), w)
		})
	})
	logger.Info(fmt.Sprintf("Listening and serving on %s", *address))
	http.ListenAndServe(*address, router)
}

// https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

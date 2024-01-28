# ai-hacking-lab
Learn about AI Hacking!

### Screenshots:

**Chatting:**

![image](https://github.com/nkcyber/ai-hacking-lab/assets/46602241/0fb0b5e3-13e7-4bf9-bebd-0791d85f193d)

## Run locally

1. Install [templ](https://templ.guide/quick-start/installation), [redis](https://redis.io/docs/install/), [Ollama](https://ollama.ai/download), and [Go](https://go.dev/doc/install).
2. ```bash
   sudo systemctl start redis # start redis
   ollama run tinyllama # install model
   # then...
   templ generate # if you've made any modifications to .templ files
   go run main.go
   ```

## Project Overview

This website is meant to serve as a framework for challenging students to manipulate large language models into doing what they want.

> [!NOTE]
> ```
> Usage of ./ai-hacking-lab:
>   -address string
>         the address to host the server on (default ":3000")
>   -maxTokens int
>         the maximum number of tokens in a response. (default 100)
>   -modelName string
>         the name of the LLM in the Ollama library (default "tinyllama")
>   -modelTemperature float
>         the 'temperature' of the LLM (default 0.1)
>   -promptPath string
>         the filepath to load prompts from (default "./example-prompts.json")
>   -redisAddress string
>         the address to connect to redis on (default "localhost:6379")
> ```
> For example,
> ```bash
> go run main.go -promptPath='/your/path/here'
> ```

### Tech Stack

| Technology                                        | Used for                |
|---------------------------------------------------|-------------------------|
| [Go](https://go.dev/)                             | Programming language    |
| [templ](https://github.com/a-h/templ)             | HTML Templating         |
| [htmx](https://htmx.org/)                         | Render chat messages    |
| [redis](https://redis.io/)                        | Store temporary chats   |
| [Ollama](https://ollama.ai/)                      | LLM access              |
| [LangChain](https://github.com/tmc/langchaingo)   | Integration with Ollama |
| [slog](https://golang.org/x/exp/slog)             | Structured Logging      |
| [go-chi](https://github.com/go-chi/chi)           | Router                  |
| [httprate](https://github.com/go-chi/httprate)    | Rate limiter            |
| [Tailwind CSS](https://tailwindcss.com/)          | CSS Framework           |
| [Hero Icons](https://github.com/tailwindlabs/heroicons)  | Icons            |


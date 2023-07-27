package server

import (
	"context"
	"encoding/json"
	"errors"
	"go-api/cache"
	"go-api/middleware"
	"go-api/structs"
	"go-api/websocket"
	"net/http"
	"os"
	"time"

	"github.com/fehernandez12/sonate"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Server struct {
	config       *structs.ServerConfig
	addr         string
	openAiClient *openai.Client
	hub          *websocket.Hub
	logger       *Logger
}

func NewServer(addr string) (*Server, error) {
	if addr == "" {
		return nil, errors.New("addr cannot be empty")
	}

	config, err := readServerConfig()
	if err != nil {
		return nil, err
	}

	cacheRep := cache.NewRedisCacheRepository()
	cache.SetRepository(cacheRep)

	return &Server{
		addr:   addr,
		config: config,
		openAiClient: openai.NewClient(
			config.OpenAIAPIKey,
		),
		hub:    websocket.NewHub(),
		logger: NewLogger(),
	}, nil
}

func (s *Server) Start(stop <-chan struct{}) error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.router(),
	}
	go s.hub.Run()
	go func() {
		logrus.WithField("addr", s.addr).Info("starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Timeout)*time.Millisecond)
	defer cancel()

	logrus.WithField("timeout", s.config.Timeout).Info("shutting down server")
	return srv.Shutdown(ctx)
}

func (s *Server) router() http.Handler {
	router := sonate.NewRouter()
	router.Use(middleware.RequestLogger)
	router.HandleFunc("/", s.defaultRoute).Methods(http.MethodPost)
	router.HandleFunc("/ws", s.hub.Handle)
	return router
}

func (s *Server) defaultRoute(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	decoder := json.NewDecoder(r.Body)
	var req structs.MessageRequest
	err := decoder.Decode(&req)
	if err != nil {
		s.logger.Error(http.StatusBadRequest, r.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	openai_req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: s.GetPreviousMessages(r.RemoteAddr),
	}
	openai_req.Messages = append(openai_req.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})
	resp, err := s.openAiClient.CreateChatCompletion(context.Background(), openai_req)
	if err != nil {
		s.logger.Error(http.StatusInternalServerError, r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	postMessage := structs.WebSocketMessage{
		Type:    "Post_Created",
		Payload: resp.Choices[0].Message.Content,
	}
	s.hub.Broadcast(postMessage, nil)
	res := &structs.MessageResponse{
		Message: resp.Choices[0].Message.Content,
	}
	response, err := json.Marshal(res)
	if err != nil {
		s.logger.Error(http.StatusInternalServerError, r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
	s.logger.Info(http.StatusCreated, r.URL.Path, start)
}

func (s *Server) GetPreviousMessages(ip string) []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage
	cached, err := cache.Get(context.Background(), ip)
	if err != nil {
		logrus.WithError(err)
		messages = make([]openai.ChatCompletionMessage, 0)
		cache.Put(context.Background(), ip, messages)
		return messages
	}
	logrus.Info(cached)
	err = json.Unmarshal([]byte(*cached), &messages)
	if err != nil {
		logrus.WithError(err)
		messages = make([]openai.ChatCompletionMessage, 0)
		cache.Put(context.Background(), ip, messages)
		return messages
	}
	return messages
}

func readServerConfig() (*structs.ServerConfig, error) {
	yamlFile, err := os.ReadFile("./server/config.yml")
	if err != nil {
		return nil, err
	}
	config := &structs.ServerConfig{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	config.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	return config, nil
}

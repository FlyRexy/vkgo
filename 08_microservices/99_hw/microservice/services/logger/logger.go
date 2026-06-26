package logger

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.vk-golang.com/vk-golang/lectures/08_microservices/99_hw/microservice/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type LogChannelRegistry struct {
	mu       *sync.RWMutex
	channels []chan *domain.Event
}

func NewChannelRegistry() *LogChannelRegistry {
	return &LogChannelRegistry{
		mu:       &sync.RWMutex{},
		channels: make([]chan *domain.Event, 0),
	}
}

func (r *LogChannelRegistry) AddChannel(c chan *domain.Event) {
	r.mu.Lock()
	r.channels = append(r.channels, c)
	r.mu.Unlock()
}

func (r *LogChannelRegistry) RemoveChannel(c chan *domain.Event) {
	newChannels := make([]chan *domain.Event, 0)
	r.mu.Lock()
	for _, channel := range r.channels {
		if channel != c {
			newChannels = append(newChannels, channel)
		}
	}
	r.channels = newChannels
	r.mu.Unlock()
}

func (r *LogChannelRegistry) Send(ev *domain.Event) {
	r.mu.RLock()
	for _, logChan := range r.channels {
		select {
		case logChan <- ev:
		default:
		}
	}
	r.mu.RUnlock()
}

type LoggerInterceptor struct {
	logRegistry *LogChannelRegistry
}

func NewLoggerInterceptor(reg *LogChannelRegistry) *LoggerInterceptor {
	return &LoggerInterceptor{
		logRegistry: reg,
	}
}

func (i LoggerInterceptor) New(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	fmt.Println("got request from ", info.FullMethod, info.Server)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("can't read metadata")
	}
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("can't read peer")
	}

	i.logRegistry.Send(&domain.Event{Timestamp: time.Now().Unix(), Consumer: md["consumer"][0], Method: info.FullMethod, Host: p.Addr.String()})

	fmt.Println(i.logRegistry.channels)

	return handler(ctx, req)
}

func (i LoggerInterceptor) Stream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	fmt.Println("got stream from ", info.FullMethod)
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("can't read metadata")
	}
	p, ok := peer.FromContext(ctx)
	if !ok {
		return errors.New("can't read peer")
	}

	i.logRegistry.Send(&domain.Event{Timestamp: time.Now().Unix(), Consumer: md["consumer"][0], Method: info.FullMethod, Host: p.Addr.String()})

	fmt.Println(i.logRegistry.channels)

	return handler(srv, ss)
}
